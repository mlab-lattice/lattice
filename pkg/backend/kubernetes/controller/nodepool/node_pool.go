package nodepool

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

const (
	finalizerName = "controller.lattice.mlab.com/node-pool"
)

func (c *Controller) syncActiveNodePool(nodePool *latticev1.NodePool) error {
	// Get the status of existing epochs.
	epochs := make(latticev1.NodePoolStatusEpochs)
	for _, epoch := range nodePool.Status.Epochs.Epochs() {
		epochInfo, ok := nodePool.Status.Epochs.Epoch(epoch)
		if !ok {
			return fmt.Errorf("could not get info for %v epoch %v", nodePool.Description(), epoch)
		}

		epochs[epoch] = latticev1.NodePoolStatusEpoch{
			NumInstances: epochInfo.NumInstances,
			InstanceType: epochInfo.InstanceType,
			State:        epochInfo.State,
		}
	}

	needsNewEpoch, err := c.cloudProvider.NodePoolNeedsNewEpoch(nodePool)
	if err != nil {
		return err
	}

	// If the node pool needs a new epoch, generate it and add it to the status.
	// Otherwise just get the current epoch and see if the cloud provider needs to act on it.
	var epoch latticev1.NodePoolEpoch
	var needsProvision bool

	if needsNewEpoch {
		epoch = nodePool.Status.Epochs.NextEpoch()
		epochs[epoch] = latticev1.NodePoolStatusEpoch{
			InstanceType: nodePool.Spec.InstanceType,
			NumInstances: nodePool.Spec.NumInstances,
			State:        latticev1.NodePoolStatePending,
		}

		needsProvision = true
	} else {
		var ok bool
		epoch, ok = nodePool.Status.Epochs.CurrentEpoch()
		if !ok {
			return fmt.Errorf("cloud provider reported that %v did not need new epoch, but it does not have a current epoch", nodePool.Description())
		}

		state, err := c.cloudProvider.NodePoolCurrentEpochState(c.latticeID, nodePool)
		if err != nil {
			return fmt.Errorf("error getting %v state for current epoch (%v) from cloud provider: %v", nodePool.Description(), epoch, err)
		}

		// Only want to call out to the cloud provider to provision the current epoch if
		// the epoch isn't already stable.
		// Due to the number of times that node pools are going to be assessed (currently have to
		// reconsider it every time any service in its namespace changes), we really want to minimize
		// the number of cloud API calls.
		needsProvision = state != latticev1.NodePoolStateStable
	}

	// Update the node pool's status prior to telling the cloud provider to provision the current epoch.
	nodePool, err = c.updateNodePoolStatus(nodePool, epochs)
	if err != nil {
		return err
	}

	// To reduce the potential number of cloud API requests, only call provision and get annotations
	// when the node pool needs to scale or be provisioned.
	if needsProvision {
		err = c.cloudProvider.ProvisionNodePoolEpoch(c.latticeID, nodePool, epoch)
		if err != nil {
			return fmt.Errorf("cloud provider could not provision %v epoch %v: %v", nodePool.Description(), epoch, err)
		}

		// Add any annotations needed by the cloud provider.
		// Copy annotations so cloud provider doesn't mutate the cache
		annotations := make(map[string]string)
		for k, v := range nodePool.Annotations {
			annotations[k] = v
		}

		err = c.cloudProvider.NodePoolAddAnnotations(c.latticeID, nodePool, annotations, epoch)
		if err != nil {
			return fmt.Errorf("cloud provider could not get annotations for %v epoch %v: %v", nodePool.Description(), epoch, err)
		}

		nodePool, err = c.updateNodePoolAnnotations(nodePool, annotations)
		if err != nil {
			return fmt.Errorf("could not update %v annotations: %v", nodePool.Description(), err)
		}
	}

	// If we got to here, the node pool's current epoch is stable, so update the status to reflect that.
	epochs[epoch] = latticev1.NodePoolStatusEpoch{
		InstanceType: nodePool.Spec.InstanceType,
		NumInstances: nodePool.Spec.NumInstances,
		State:        latticev1.NodePoolStateStable,
	}
	nodePool, err = c.updateNodePoolStatus(nodePool, epochs)
	if err != nil {
		return err
	}

	// We successfully provisioned the current epoch, so we can start trying to retire
	// any earlier epochs.
	_, err = c.retireEpochs(nodePool, false)
	return err
}

func (c *Controller) syncDeletedNodePool(nodePool *latticev1.NodePool) error {
	nodePool, err := c.retireEpochs(nodePool, true)
	if err != nil {
		return err
	}

	// If there are still epochs that weren't able to be retired yet, we are done for this round.
	// Workloads will be terminated and this node pool will be requeued and reassessed.
	if len(nodePool.Status.Epochs) > 0 {
		return nil
	}

	// If all of the epochs were retired, the node pool has been deleted so we can remove the finalizer.
	_, err = c.removeFinalizer(nodePool)
	return err
}

func (c *Controller) retireEpochs(nodePool *latticev1.NodePool, retireCurrent bool) (*latticev1.NodePool, error) {
	retiredEpochs := set.NewSet()
	for _, epoch := range nodePool.Status.Epochs.Epochs() {
		if !retireCurrent {
			currentEpoch, ok := nodePool.Status.Epochs.CurrentEpoch()
			if !ok {
				return nil, fmt.Errorf("trying to retire %v epochs but it does not have a current epoch", nodePool.Description())
			}

			if epoch == currentEpoch {
				continue
			}
		}

		// If the node pool can be retired, ask the cloud provider to deprovision it.
		retired, err := c.isEpochRetired(nodePool, epoch)
		if err != nil {
			return nil, fmt.Errorf("error trying to check if %v epoch %v is retired: %v", nodePool.Description(), epoch, err)
		}

		if !retired {
			continue
		}

		err = c.cloudProvider.DeprovisionNodePoolEpoch(c.latticeID, nodePool, epoch)
		if err != nil {
			return nil, fmt.Errorf("cloud provider could not deprovision %v epoch %v: %v", nodePool.Description(), epoch, err)
		}

		retiredEpochs.Add(epoch)
	}

	// Create a new map of epochs for the node pool's status, removing the epochs that were just retired.
	epochs := make(map[latticev1.NodePoolEpoch]latticev1.NodePoolStatusEpoch)
	for epoch, epochInfo := range nodePool.Status.Epochs {
		if retiredEpochs.Contains(epoch) {
			continue
		}

		epochs[epoch] = epochInfo
	}

	return c.updateNodePoolStatus(nodePool, epochs)
}

func (c *Controller) isEpochRetired(nodePool *latticev1.NodePool, epoch latticev1.NodePoolEpoch) (bool, error) {
	// Check to see if any workload is still possibly using the epoch
	serviceRunning, err := c.serviceRunningOnEpoch(nodePool, epoch)
	if err != nil {
		return false, err
	}

	if serviceRunning {
		return false, nil
	}

	// TODO: check for jobs when we have them

	return true, nil
}

func (c *Controller) serviceRunningOnEpoch(nodePool *latticev1.NodePool, epoch latticev1.NodePoolEpoch) (bool, error) {
	// Check to see if any workload is still possibly using the epoch
	services, err := c.serviceLister.Services(nodePool.Namespace).List(labels.Everything())
	if err != nil {
		return false, err
	}

	for _, service := range services {
		annotation, err := service.NodePoolAnnotation()
		if err != nil {
			// FIXME: send warning
			// If we returned an error here then one service with an invalid annotation
			// would make a node pool permanently error out.
			// So we have to decide whether the epoch is retired if there is only a service
			// with an invalid annotation.
			// In order to make forward progress, we won't prevent an epoch from being retired
			// if there's an invalid annotation.
			continue
		}

		// If the service hasn't annotated its intended node pool yet, it's possible it's in
		// the process of annotating it with this epoch, so it's possible there's a service
		// running on the epoch.
		if annotation.IsEmpty() {
			return true, nil
		}

		if annotation.ContainsEpoch(nodePool.Namespace, nodePool.Name, epoch) {
			return true, nil
		}

		// If the service is at least partially running on a larger epoch of the node pool
		// and is not running on this epoch of the node pool, then it will never run on this
		// epoch.
		if annotation.ContainsLargerEpoch(nodePool.Namespace, nodePool.Name, epoch) {
			continue
		}

		// If the service controller hasn't processed the most recent update to the service,
		// it's possible that it's in the process of annotating it with this epoch.
		if !service.UpdateProcessed() {
			return true, nil
		}

		// Otherwise, the service controller has seen and reacted to this version of the service.
		// If the service were updated to run on this node pool, the service controller would have
		// at least this current version of the node pool in its cache, if not a more recent version,
		// and therefore would not assign the service to this epoch.
		// NOTE: this relies on the fact that the nodepool controller and service controller are both
		// operating on the same cache.
		continue
	}

	// Didn't find any services that could be running on the epoch.
	return false, nil
}

func (c *Controller) updateNodePoolAnnotations(nodePool *latticev1.NodePool, annotations map[string]string) (*latticev1.NodePool, error) {
	if reflect.DeepEqual(nodePool.Annotations, annotations) {
		return nodePool, nil
	}

	// Copy so we don't mutate the cache
	nodePool = nodePool.DeepCopy()
	nodePool.Annotations = annotations

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) updateNodePoolStatus(
	nodePool *latticev1.NodePool,
	epochs map[latticev1.NodePoolEpoch]latticev1.NodePoolStatusEpoch,
) (*latticev1.NodePool, error) {
	state, err := nodePoolState(epochs)
	if err != nil {
		return nil, fmt.Errorf("error trying to get state for %v: %v", nodePool.Description(), err)
	}

	status := latticev1.NodePoolStatus{
		ObservedGeneration: nodePool.ObjectMeta.Generation,
		State:              state,
		Epochs:             epochs,
	}

	if reflect.DeepEqual(nodePool.Status, status) {
		return nodePool, nil
	}

	// Copy the service so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Status = status

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).UpdateStatus(nodePool)
}

func nodePoolState(epochs latticev1.NodePoolStatusEpochs) (latticev1.NodePoolState, error) {
	if len(epochs) == 0 {
		return latticev1.NodePoolStatePending, nil
	}

	if len(epochs) > 1 {
		return latticev1.NodePoolStateUpdating, nil
	}

	current, ok := epochs.CurrentEpoch()
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("epochs had an epoch but could not get current epoch")
	}

	epochInfo, ok := epochs.Epoch(current)
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("epochs had a current epoch but could not get current epoch info")
	}

	return epochInfo.State, nil
}

func (c *Controller) addFinalizer(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range nodePool.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("NodePool %v has %v finalizer", nodePool.Name, finalizerName)
			return nodePool, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	nodePool.Finalizers = append(nodePool.Finalizers, finalizerName)
	glog.V(5).Infof("NodePool %v missing %v finalizer, adding it", nodePool.Name, finalizerName)

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) removeFinalizer(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range nodePool.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return nodePool, nil
	}

	// The finalizer was in the list, so we should remove it.
	nodePool.Finalizers = finalizers
	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}
