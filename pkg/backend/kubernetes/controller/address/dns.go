package address

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"reflect"
)

func (c *Controller) syncDNS(address *latticev1.Address) error {
	endpoint, err := c.endpointLister.Endpoints(address.Namespace).Get(address.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createNewEndpoint(address)
		}

		return nil, err
	}

	endpoint, err = c.syncExistingEndpoint(address, endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (c *Controller) syncExistingEndpoint(
	address *latticev1.Address,
	endpoint *latticev1.Endpoint,
) (*latticev1.Endpoint, error) {
	spec, err := c.serviceMesh.GetEndpointSpec(address)
	if err != nil {
		return nil, err
	}

	return c.updateEndpointSpec(endpoint, *spec)
}

func (c *Controller) updateEndpointSpec(
	endpoint *latticev1.Endpoint,
	desiredSpec latticev1.EndpointSpec,
) (*latticev1.Endpoint, error) {
	if reflect.DeepEqual(endpoint.Spec, desiredSpec) {
		return endpoint, nil
	}

	// Copy so the shared cache isn't mutated
	endpoint = endpoint.DeepCopy()
	endpoint.Spec = desiredSpec

	return c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
}

func (c *Controller) createNewEndpoint(address *latticev1.Address) (*latticev1.Endpoint, error) {
	endpoint, err := c.newEndpoint(address)
	if err != nil {
		return nil, err
	}

	endpoint, err = c.latticeClient.LatticeV1().Endpoints(address.Namespace).Create(endpoint)
	if err != nil {
		// FIXME: check for AlreadyExists/Conflict
		return nil, err
	}

	return endpoint, nil
}

func (c *Controller) newEndpoint(address *latticev1.Address) (*latticev1.Endpoint, error) {
	spec, err := c.serviceMesh.GetEndpointSpec(address)
	if err != nil {
		return nil, err
	}

	endpoint := &latticev1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:            address.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(address, controllerKind)},
		},
		Spec: *spec,
		Status: latticev1.EndpointStatus{
			State: latticev1.EndpointStatePending,
		},
	}

	return endpoint, nil
}
