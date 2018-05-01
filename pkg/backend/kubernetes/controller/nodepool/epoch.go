package nodepool

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

func (c *Controller) retireEpochs(nodePool *latticev1.NodePool, retireCurrent bool) (*latticev1.NodePool, error) {
	retiredEpochs := mapset.NewSet()
	for _, epoch := range nodePool.Status.Epochs.Epochs() {
		if !retireCurrent {
			currentEpoch, ok := nodePool.Status.Epochs.CurrentEpoch()
			if !ok {
				return nil, fmt.Errorf("trying to retire %v epochs but it does not have a current epoch", nodePool.Description(c.namespacePrefix))
			}

			if epoch == currentEpoch {
				continue
			}
		}

		// If the node pool can be retireable, ask the cloud provider to deprovision it.
		retireable, reason, err := c.isEpochRetired(nodePool, epoch)
		if err != nil {
			return nil, fmt.Errorf("error trying to check if %v epoch %v is retireable: %v", nodePool.Description(c.namespacePrefix), epoch, err)
		}

		if !retireable {
			glog.V(4).Infof("%v epoch %v is not able to be retireable: %v", nodePool.Description(c.namespacePrefix), epoch, reason)
			continue
		}

		err = c.cloudProvider.DestroyNodePoolEpoch(c.latticeID, nodePool, epoch)
		if err != nil {
			return nil, fmt.Errorf("cloud provider could not deprovision %v epoch %v: %v", nodePool.Description(c.namespacePrefix), epoch, err)
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

func (c *Controller) isEpochRetired(nodePool *latticev1.NodePool, epoch latticev1.NodePoolEpoch) (bool, string, error) {
	// Check to see if any workload is still possibly using the epoch
	serviceRunning, err := c.serviceRunningOnEpoch(nodePool, epoch)
	if err != nil {
		return false, "", err
	}

	if serviceRunning {
		return false, "services still potentially running", nil
	}

	// TODO: check for jobs when we have them

	return true, "", nil
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

		// If the service is being deleted, if it contains the epoch then the service is
		// clearly possibly still running on the epoch. But if it doesn't have the epoch
		// annotated we know that it will not in the future, so we are safe to say
		// the service will not be running on this epoch.
		if service.Deleted() {
			if annotation.ContainsEpoch(nodePool.Namespace, nodePool.Name, epoch) {
				return true, nil
			}

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
