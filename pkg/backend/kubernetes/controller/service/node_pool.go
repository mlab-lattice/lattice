package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncServiceNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	// TODO(kevinrosendahl): add support for shared node pools
	nodePool, err := c.nodePoolLister.NodePools(service.Namespace).Get(service.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createNewNodePool(service)
		}

		return nil, err
	}

	nodePool, err = c.syncExistingNodePool(service, nodePool)
	if err != nil {
		return nil, err
	}

	return nodePool, nil
}

func (c *Controller) syncExistingNodePool(service *latticev1.Service, nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// TODO(kevinrosendahl): only change NodePool spec for dedicated node pools
	spec, err := nodePoolSpec(service)
	if err != nil {
		return nil, err
	}

	if spec.InstanceType != nodePool.Spec.InstanceType {
		glog.V(4).Infof("NodePool %v for Service %v/%v had out of date instance type, updating", nodePool.Name, service.Namespace, service.Name)
		return c.updateNodePoolSpec(nodePool, spec)
	}

	if spec.NumInstances != nodePool.Spec.NumInstances {
		glog.V(4).Infof("NodePool %v for Service %v/%v had out of date num instances, updating", nodePool.Name, service.Namespace, service.Name)
		return c.updateNodePoolSpec(nodePool, spec)
	}

	return nodePool, nil
}

func (c *Controller) updateNodePoolSpec(nodePool *latticev1.NodePool, desiredSpec latticev1.NodePoolSpec) (*latticev1.NodePool, error) {
	// Copy so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Spec = desiredSpec

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) createNewNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	nodePool, err := newNodePool(service)
	if err != nil {
		return nil, err
	}

	nodePool, err = c.latticeClient.LatticeV1().NodePools(service.Namespace).Create(nodePool)
	if err != nil {
		return nil, err
	}

	return nodePool, nil
}

func newNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	spec, err := nodePoolSpec(service)
	if err != nil {
		return nil, err
	}

	nodePool := &latticev1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: spec,
		Status: latticev1.NodePoolStatus{
			State: latticev1.NodePoolStatePending,
		},
	}

	return nodePool, nil
}

func nodePoolSpec(service *latticev1.Service) (latticev1.NodePoolSpec, error) {
	if service.Spec.Definition.Resources().InstanceType == nil {
		return latticev1.NodePoolSpec{}, fmt.Errorf("cannot create NodePool for Service with no resources.instance_type")
	}
	instanceType := *service.Spec.Definition.Resources().InstanceType

	var numInstances int32
	if service.Spec.Definition.Resources().NumInstances != nil {
		numInstances = *service.Spec.Definition.Resources().NumInstances
	} else if service.Spec.Definition.Resources().MinInstances != nil {
		numInstances = *service.Spec.Definition.Resources().MinInstances
	} else {
		return latticev1.NodePoolSpec{}, fmt.Errorf("cannot create NodePool for Service with neither resources.num_instances nor resources.min_instances")
	}

	spec := latticev1.NodePoolSpec{
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
	return spec, nil
}
