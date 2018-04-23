package address

import (
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

const (
	finalizerName = "controller.lattice.mlab.com/address"
)

func (c *Controller) updateAddressStatus(
	address *latticev1.Address,
	state latticev1.AddressState,
	failureInfo *latticev1.AddressStatusFailureInfo,
	ports map[int32]string,
) (*latticev1.Address, error) {
	status := latticev1.AddressStatus{
		ObservedGeneration: address.Generation,
		State:              state,
		FailureInfo:        failureInfo,
		Ports:              ports,
	}

	if reflect.DeepEqual(address.Status, status) {
		return address, nil
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Status = status

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).UpdateStatus(address)
}

func (c *Controller) updateAddressAnnotations(address *latticev1.Address, annotations map[string]string) (*latticev1.Address, error) {
	if reflect.DeepEqual(address.Annotations, annotations) {
		return address, nil
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Annotations = annotations

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
}

func (c *Controller) addFinalizer(address *latticev1.Address) (*latticev1.Address, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range address.Finalizers {
		if finalizer == finalizerName {
			return address, nil
		}
	}

	// Copy so we don't mutate the shared cache
	address = address.DeepCopy()
	address.Finalizers = append(address.Finalizers, finalizerName)

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
}

func (c *Controller) removeFinalizer(address *latticev1.Address) (*latticev1.Address, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range address.Finalizers {
		if finalizer == finalizerName {
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

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
}
