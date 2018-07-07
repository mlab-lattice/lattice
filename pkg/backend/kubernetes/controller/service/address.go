package service

import (
	"fmt"
	//"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncAddress(service *latticev1.Service) (*latticev1.Address, error) {
	// GEB: should endpoints be synced here instead of within the address controller?
	address, err := c.address(service)
	if err != nil {
		return nil, err
	}

	if address == nil {
		address, err = c.createNewAddress(service)
		if err != nil {
			return nil, err
		}
	}

	address, err = c.syncExistingAddress(service, address)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (c *Controller) syncExistingAddress(service *latticev1.Service, address *latticev1.Address) (*latticev1.Address, error) {
	// GEB: this does not set Endpoints, so updateAddressSpec will always result in an update
	spec, err := c.addressSpec(service)
	if err != nil {
		return nil, err
	}

	return c.updateAddressSpec(address, spec)
}

func (c *Controller) updateAddressSpec(address *latticev1.Address, spec latticev1.AddressSpec) (*latticev1.Address, error) {
	// GEB: this does not take ExternalAddress or Endpoints into account... if either of these are update
	//      in another controller, we go into a loop volleying update back and forth between controllers.
	//      how best to handle this?
	// if reflect.DeepEqual(address.Spec, spec) {
	// 	return address, nil
	// }

	if *address.Spec.Service == *spec.Service {
		return address, nil
	}

	// Copy so the shared cache isn't mutated
	address = address.DeepCopy()
	address.Spec = spec

	result, err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
	if err != nil {
		err := fmt.Errorf("error trying to update %v: %v", address.Description(c.namespacePrefix), err)
		return nil, err
	}

	return result, nil
}

func (c *Controller) createNewAddress(service *latticev1.Service) (*latticev1.Address, error) {
	serviceAddress, err := c.newServiceAddress(service)
	if err != nil {
		return nil, err
	}

	result, err := c.latticeClient.LatticeV1().Addresses(service.Namespace).Create(serviceAddress)
	if err != nil {
		err := fmt.Errorf("error trying to create new address for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	return result, nil
}

func (c *Controller) newServiceAddress(service *latticev1.Service) (*latticev1.Address, error) {
	spec, err := c.addressSpec(service)
	if err != nil {
		return nil, err
	}

	path, err := service.PathLabel()
	if err != nil {
		err := fmt.Errorf("error getting %v path label: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	serviceAddress := &latticev1.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name:            service.Name,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(service)},
			Labels: map[string]string{
				latticev1.AddressPathLabelKey: path.ToDomain(),
				latticev1.ServiceIDLabelKey:   service.Name,
			},
		},
		Spec: spec,
	}

	return serviceAddress, nil
}

func (c *Controller) addressSpec(service *latticev1.Service) (latticev1.AddressSpec, error) {
	path, err := service.PathLabel()
	if err != nil {
		err := fmt.Errorf("error getting %v path label: %v", service.Description(c.namespacePrefix), err)
		return latticev1.AddressSpec{}, err
	}

	spec := latticev1.AddressSpec{
		Service: &path,
	}
	return spec, nil
}

func (c *Controller) address(service *latticev1.Service) (*latticev1.Address, error) {
	address, err := c.cachedAddress(service)
	if err != nil {
		return nil, err
	}

	if address != nil {
		return address, nil
	}

	return c.quorumAddress(service)
}

func (c *Controller) cachedAddress(service *latticev1.Service) (*latticev1.Address, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	addresses, err := c.addressLister.Addresses(service.Namespace).List(selector)
	if err != nil {
		err := fmt.Errorf("error tyring to get cached address for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, nil
	}

	if len(addresses) > 1 {
		return nil, fmt.Errorf("found multiple cached addresses for %v", service.Description(c.namespacePrefix))
	}

	return addresses[0], nil
}

func (c *Controller) quorumAddress(service *latticev1.Service) (*latticev1.Address, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	addressList, err := c.latticeClient.LatticeV1().Addresses(service.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		err := fmt.Errorf("error tyring to get address for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	addresses := addressList.Items
	if len(addresses) == 0 {
		return nil, nil
	}

	if len(addresses) > 1 {
		return nil, fmt.Errorf("found multiple addresses for %v", service.Description(c.namespacePrefix))
	}

	return &addresses[0], nil
}
