package serviceaddress

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"reflect"
)

func (c *Controller) syncEndpoint(address *crv1.ServiceAddress) (*crv1.Endpoint, error) {
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

func (c *Controller) syncExistingEndpoint(address *crv1.ServiceAddress, endpoint *crv1.Endpoint) (*crv1.Endpoint, error) {
	spec, err := c.serviceMesh.GetEndpointSpec(address)
	if err != nil {
		return nil, err
	}

	return c.updateEndpointSpec(endpoint, *spec)
}

func (c *Controller) updateEndpointSpec(endpoint *crv1.Endpoint, desiredSpec crv1.EndpointSpec) (*crv1.Endpoint, error) {
	if reflect.DeepEqual(endpoint.Spec, desiredSpec) {
		return endpoint, nil
	}

	// Copy so the shared cache isn't mutated
	endpoint = endpoint.DeepCopy()
	endpoint.Spec = desiredSpec

	return c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
}

func (c *Controller) createNewEndpoint(address *crv1.ServiceAddress) (*crv1.Endpoint, error) {
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

func (c *Controller) newEndpoint(address *crv1.ServiceAddress) (*crv1.Endpoint, error) {
	spec, err := c.serviceMesh.GetEndpointSpec(address)
	if err != nil {
		return nil, err
	}

	endpoint := &crv1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:            address.Name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(address, controllerKind)},
		},
		Spec: *spec,
		Status: crv1.EndpointStatus{
			State: crv1.EndpointStatePending,
		},
	}

	return endpoint, nil
}
