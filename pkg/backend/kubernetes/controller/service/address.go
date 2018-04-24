package service

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncAddress(service *latticev1.Service) (*latticev1.Address, error) {
	path, err := service.PathLabel()
	if err != nil {
		return nil, fmt.Errorf("error getting path for %v: %v", service.Description(c.namespacePrefix), err)
	}

	address, err := c.address(service.Namespace, path)
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
	spec, err := c.addressSpec(service)
	if err != nil {
		return nil, err
	}

	return c.updateAddressSpec(address, spec)
}

func (c *Controller) updateAddressSpec(address *latticev1.Address, spec latticev1.AddressSpec) (*latticev1.Address, error) {
	if reflect.DeepEqual(address.Spec, spec) {
		return address, nil
	}

	// Copy so the shared cache isn't mutated
	address = address.DeepCopy()
	address.Spec = spec

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)
}

func (c *Controller) createNewAddress(service *latticev1.Service) (*latticev1.Address, error) {
	serviceAddress, err := c.newServiceAddress(service)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().Addresses(service.Namespace).Create(serviceAddress)
}

func (c *Controller) newServiceAddress(service *latticev1.Service) (*latticev1.Address, error) {
	spec, err := c.addressSpec(service)
	if err != nil {
		return nil, err
	}

	serviceAddress := &latticev1.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: spec,
		Status: latticev1.AddressStatus{
			State: latticev1.AddressStateUpdating,
		},
	}

	return serviceAddress, nil
}

func (c *Controller) addressSpec(service *latticev1.Service) (latticev1.AddressSpec, error) {
	path, err := service.PathLabel()
	if err != nil {
		return latticev1.AddressSpec{}, err
	}

	spec := latticev1.AddressSpec{
		Service: &path,
	}
	return spec, nil
}

func (c *Controller) address(namespace string, path tree.NodePath) (*latticev1.Address, error) {
	address, err := c.cachedAddress(namespace, path)
	if err != nil {
		return nil, err
	}

	if address != nil {
		return address, nil
	}

	return c.quorumAddress(namespace, path)
}

func (c *Controller) cachedAddress(namespace string, path tree.NodePath) (*latticev1.Address, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.AddressPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	addresses, err := c.addressLister.Addresses(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, nil
	}

	if len(addresses) > 1 {
		return nil, fmt.Errorf("found multiple addresses for path %v in namespace %v", path.String(), namespace)
	}

	return addresses[0], nil
}

func (c *Controller) quorumAddress(namespace string, path tree.NodePath) (*latticev1.Address, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.AddressPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	addressList, err := c.latticeClient.LatticeV1().Addresses(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	addresses := addressList.Items
	if len(addresses) == 0 {
		return nil, nil
	}

	if len(addresses) > 1 {
		return nil, fmt.Errorf("found multiple addresses for path %v in namespace %v", path.String(), namespace)
	}

	return &addresses[0], nil
}
