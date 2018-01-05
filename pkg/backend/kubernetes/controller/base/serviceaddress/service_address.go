package serviceaddress

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncServiceAddressStatus(address *crv1.ServiceAddress, endpoint *crv1.Endpoint) (*crv1.ServiceAddress, error) {
	var state crv1.ServiceAddressState
	switch endpoint.Status.State {
	case crv1.EndpointStatePending:
		state = crv1.ServiceAddressStatePending

	case crv1.EndpointStateCreated:
		state = crv1.ServiceAddressStateCreated

	case crv1.EndpointStateFailed:
		state = crv1.ServiceAddressStateFailed

	default:
		return nil, fmt.Errorf("Endpoint %v/%v in unexpected state %v", endpoint.Namespace, endpoint.Name, endpoint.Status.State)
	}

	return c.updateServiceStatus(address, state)
}

func (c *Controller) updateServiceStatus(address *crv1.ServiceAddress, state crv1.ServiceAddressState) (*crv1.ServiceAddress, error) {
	status := crv1.ServiceAddressStatus{
		State:              state,
		ObservedGeneration: address.Generation,
	}

	if reflect.DeepEqual(address.Status, status) {
		return address, nil
	}

	// Copy the address so the shared cache isn't mutated
	address = address.DeepCopy()
	address.Status = status

	return c.latticeClient.LatticeV1().ServiceAddresses(address.Namespace).Update(address)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().Services(address.Namespace).UpdateStatus(address)
}
