package service

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

func (c *Controller) syncServiceNodePool(service *crv1.Service) (*crv1.Service, *crv1.NodePool, error) {
	nodePoolID, ok := service.Labels[kubeconstants.LabelKeyNodePoolID]
	if !ok {
		// TODO: add support here for shared node pools
		return c.createNewNodePool(service)
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(nodePoolID)
	if err != nil {
		return nil, nil, err
	}

	nodePool, err := c.nodePoolLister.NodePools(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createNewNodePool(service)
		}

		return nil, nil, err
	}

	nodePool, err = c.syncExistingNodePool(service, nodePool)
	if err != nil {
		return nil, nil, err
	}

	return service, nodePool, nil
}

func (c *Controller) syncExistingNodePool(service *crv1.Service, nodePool *crv1.NodePool) (*crv1.NodePool, error) {
	// TODO: only change NodePool spec for dedicated node pools
	desiredSpec, err := nodePoolSpec(service)
	if err != nil {
		return nil, err
	}

	if desiredSpec.InstanceType != nodePool.Spec.InstanceType {
		glog.V(4).Infof("NodePool %v for Service %v/%v had out of date instance type, updating", nodePool.Name, service.Namespace, service.Name)
		return c.updateNodePoolSpec(nodePool, desiredSpec)
	}

	if desiredSpec.NumInstances != nodePool.Spec.NumInstances {
		glog.V(4).Infof("NodePool %v for Service %v/%v had out of date num instances, updating", nodePool.Name, service.Namespace, service.Name)
		return c.updateNodePoolSpec(nodePool, desiredSpec)
	}

	return nodePool, nil
}

func (c *Controller) updateNodePoolSpec(nodePool *crv1.NodePool, desiredSpec *crv1.NodePoolSpec) (*crv1.NodePool, error) {
	nodePool.Spec = *desiredSpec
	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) createNewNodePool(service *crv1.Service) (*crv1.Service, *crv1.NodePool, error) {
	nodePool, err := c.nodePoolLister.NodePools(service.Namespace).Get(service.Name)
	if err == nil {
		// If the node pool already exists, then the Service simply hasn't been labelled with it.
		service, err := c.addNodePoolLabel(service, nodePool)
		if err != nil {
			return nil, nil, err
		}
		return service, nodePool, nil
	}

	if err != nil && errors.IsNotFound(err) {
		return nil, nil, err
	}

	nodePool, err = newNodePool(service)
	if err != nil {
		return nil, nil, err
	}

	nodePool, err = c.latticeClient.LatticeV1().NodePools(service.Namespace).Create(nodePool)
	if err != nil {
		return nil, nil, err
	}

	service, err = c.addNodePoolLabel(service, nodePool)
	if err != nil {
		return nil, nil, err
	}

	return service, nodePool, nil
}

func (c *Controller) addNodePoolLabel(service *crv1.Service, nodePool *crv1.NodePool) (*crv1.Service, error) {
	serviceCopy := service.DeepCopy()
	serviceCopy.Labels[kubeconstants.LabelKeyNodePoolID] = kubeutil.NodePoolIDLabelValue(nodePool)
	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}

func newNodePool(service *crv1.Service) (*crv1.NodePool, error) {
	spec, err := nodePoolSpec(service)
	if err != nil {
		return nil, err
	}

	nodePool := &crv1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: *spec,
		Status: crv1.NodePoolStatus{
			State: crv1.NodePoolStatePending,
		},
	}

	return nodePool, nil
}

func nodePoolSpec(service *crv1.Service) (*crv1.NodePoolSpec, error) {
	if service.Spec.Definition.Resources.InstanceType == nil {
		return nil, fmt.Errorf("cannot create NodePool for Service with no resources.instance_type")
	}
	instanceType := *service.Spec.Definition.Resources.InstanceType

	var numInstances int32
	if service.Spec.Definition.Resources.NumInstances != nil {
		numInstances = *service.Spec.Definition.Resources.NumInstances
	} else if service.Spec.Definition.Resources.MinInstances != nil {
		numInstances = *service.Spec.Definition.Resources.MinInstances
	} else {
		return nil, fmt.Errorf("cannot create NodePool for Service with neither resources.num_instances nor resources.min_instances")
	}

	spec := &crv1.NodePoolSpec{
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
	return spec, nil
}
