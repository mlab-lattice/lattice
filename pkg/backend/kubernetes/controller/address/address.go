package address

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
)

func (c *Controller) updateAddressStatus(
	address *latticev1.Address,
	state latticev1.AddressState,
	message *string,
	failureInfo *latticev1.AddressStatusFailureInfo,
	ports map[int32]string,
) (*latticev1.Address, error) {
	status := latticev1.AddressStatus{
		ObservedGeneration: address.Generation,

		State:       state,
		Message:     message,
		FailureInfo: failureInfo,

		Ports: ports,
	}

	if reflect.DeepEqual(address.Status, status) {
		return address, nil
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Status = status

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).UpdateStatus(address)
	if err != nil {
		return nil, fmt.Errorf("error updating %v status: %v", address.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) updateAddressEndpoints(
	address *latticev1.Address,
	endpoints []string,
) (*latticev1.Address, error) {
	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Spec.Endpoints = endpoints

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
	if err != nil {
		return nil, fmt.Errorf("error updating %v endpoints: %v", address.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) updateAddressAnnotations(address *latticev1.Address, annotations map[string]string) (*latticev1.Address, error) {
	if reflect.DeepEqual(address.Annotations, annotations) {
		return address, nil
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Annotations = annotations

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
	if err != nil {
		return nil, fmt.Errorf("error updating %v annotations: %v", address.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) addFinalizer(address *latticev1.Address) (*latticev1.Address, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range address.Finalizers {
		if finalizer == kubeutil.AddressControllerFinalizer {
			return address, nil
		}
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Finalizers = append(address.Finalizers, kubeutil.AddressControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", address.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(address *latticev1.Address) (*latticev1.Address, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range address.Finalizers {
		if finalizer == kubeutil.AddressControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return address, nil
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", address.Description(c.namespacePrefix), err)
	}

	return result, nil
}
