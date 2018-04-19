package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncLoadBalancer(service *latticev1.Service, nodePool *latticev1.NodePool, nodePoolReady bool) (*latticev1.LoadBalancer, bool, error) {
	lbNeeded := needsLoadBalancer(service)

	if !nodePoolReady {
		return nil, lbNeeded, nil
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			if !lbNeeded {
				return nil, false, nil
			}

			loadBalancer, err := c.createNewLoadBalancer(service, nodePool)
			return loadBalancer, true, err
		}

		return nil, lbNeeded, err
	}

	return loadBalancer, lbNeeded, nil
}

func needsLoadBalancer(service *latticev1.Service) bool {
	for _, ports := range service.Spec.Ports {
		for _, port := range ports {
			if port.Public {
				return true
			}
		}
	}

	return false
}

func (c *Controller) createNewLoadBalancer(service *latticev1.Service, nodePool *latticev1.NodePool) (*latticev1.LoadBalancer, error) {
	loadBalancer := newLoadBalancer(service, nodePool)

	loadBalancer, err := c.latticeClient.LatticeV1().LoadBalancers(service.Namespace).Create(loadBalancer)
	if err != nil {
		// FIXME: check for AlreadyExists/Conflict
		return nil, err
	}

	return loadBalancer, nil
}

func newLoadBalancer(service *latticev1.Service, nodePool *latticev1.NodePool) *latticev1.LoadBalancer {
	return &latticev1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            service.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(service, controllerKind)},
		},
		Spec: latticev1.LoadBalancerSpec{
			NodePool: nodePool.Name,
		},
		Status: latticev1.LoadBalancerStatus{
			State: latticev1.LoadBalancerStatePending,
		},
	}
}
