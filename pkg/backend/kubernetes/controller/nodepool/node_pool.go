package nodepool

import (
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/golang/glog"
)

const (
	finalizerName = "controller.lattice.mlab.com/node-pool"
)

func (c *Controller) syncActiveNodePool(nodePool *latticev1.NodePool) error {
	state, err := c.cloudProvider.NodePoolState(c.latticeID, nodePool)
	if err != nil {
		return err
	}

	nodePool, err = c.updateNodePoolStatus(nodePool, state)
	if err != nil {
		return err
	}

	nodePool, requeueTime, err := c.cloudProvider.ProvisionNodePool(c.latticeID, nodePool)
	if err != nil {
		return err
	}

	nodePool, err = c.updateNodePool(nodePool)
	if err != nil {
		return err
	}

	if requeueTime != nil {
		c.queue.AddAfter(nodePool, *requeueTime)
		return nil
	}

	// FIXME: drain then deprovision old nodes
	return nil
}

func (c *Controller) syncDeletedNodePool(nodePool *latticev1.NodePool) error {
	// FIXME: add drain
	requeueTime, err := c.cloudProvider.DeprovisionNodePool(c.latticeID, nodePool)
	if err != nil {
		return err
	}

	if requeueTime != nil {
		c.queue.AddAfter(nodePool, *requeueTime)
		return nil
	}

	_, err = c.removeFinalizer(nodePool)
	return err
}

func (c *Controller) updateNodePool(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) updateNodePoolStatus(nodePool *latticev1.NodePool, state latticev1.NodePoolState) (*latticev1.NodePool, error) {
	status := latticev1.NodePoolStatus{
		ObservedGeneration: nodePool.ObjectMeta.Generation,
		State:              state,
	}

	if reflect.DeepEqual(nodePool.Status, status) {
		return nodePool, nil
	}

	// Copy the service so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Status = status

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).UpdateStatus(nodePool)
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
