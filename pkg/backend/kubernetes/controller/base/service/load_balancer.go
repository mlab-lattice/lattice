package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncLoadBalancer(service *crv1.Service) (*crv1.LoadBalancer, bool, error) {
	lbNeeded := needsLoadBalancer(service)

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			if !lbNeeded {
				return nil, false, nil
			}

			loadBalancer, err := c.createNewLoadBalancer(service)
			return loadBalancer, true, err
		}

		return nil, lbNeeded, err
	}

	return loadBalancer, lbNeeded, nil
}

func needsLoadBalancer(service *crv1.Service) bool {
	for _, ports := range service.Spec.Ports {
		for _, port := range ports {
			if port.Public {
				return true
			}
		}
	}

	return false
}

func (c *Controller) createNewLoadBalancer(service *crv1.Service) (*crv1.LoadBalancer, error) {
	loadBalancer := newLoadBalancer(service)

	loadBalancer, err := c.latticeClient.LatticeV1().LoadBalancers(service.Namespace).Create(loadBalancer)
	if err != nil {
		// FIXME: check for AlreadyExists/Conflict
		return nil, err
	}

	return loadBalancer, nil
}

func newLoadBalancer(service *crv1.Service) *crv1.LoadBalancer {
	return &crv1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            service.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(service, controllerKind)},
		},
		Spec: crv1.LoadBalancerSpec{},
		Status: crv1.LoadBalancerStatus{
			State: crv1.LoadBalancerStatePending,
		},
	}
}
