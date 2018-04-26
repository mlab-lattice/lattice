package nodepool

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
)

func (c *Controller) updateNodePoolAnnotations(nodePool *latticev1.NodePool, annotations map[string]string) (*latticev1.NodePool, error) {
	if reflect.DeepEqual(nodePool.Annotations, annotations) {
		return nodePool, nil
	}

	// Copy so we don't mutate the cache
	nodePool = nodePool.DeepCopy()
	nodePool.Annotations = annotations

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error updating annotations for %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) updateNodePoolStatus(
	nodePool *latticev1.NodePool,
	epochs map[latticev1.NodePoolEpoch]latticev1.NodePoolStatusEpoch,
) (*latticev1.NodePool, error) {
	state, err := nodePoolState(epochs)
	if err != nil {
		return nil, fmt.Errorf("error trying to get state for %v: %v", nodePool.Description(c.namespacePrefix), err)
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

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).UpdateStatus(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error trying to update %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return result, nil
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
		if finalizer == kubeutil.NodePoolControllerFinalizer {
			return nodePool, nil
		}
	}

	// Copy so we don't mutate the shared cache
	nodePool = nodePool.DeepCopy()
	nodePool.Finalizers = append(nodePool.Finalizers, kubeutil.NodePoolControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error adding finalizer to %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range nodePool.Finalizers {
		if finalizer == kubeutil.NodePoolControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return nodePool, nil
	}

	// Copy so we don't mutate the shared cache
	nodePool = nodePool.DeepCopy()
	nodePool.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
	if err != nil {
		return nil, fmt.Errorf("error removing finalizer from %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	return result, nil
}
