package service

import (
	"reflect"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/block"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncServiceServiceAddress(service *latticev1.Service) (*latticev1.ServiceAddress, error) {
	serviceAddress, err := c.serviceAddressLister.ServiceAddresses(service.Namespace).Get(service.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// ServiceAddress not found in our cache, but its possible it could still have been created,
			// check the API to see it exists.
			// TODO: would it be better to just try to create it and check for a Conflict? Could save a round of queries per new rollout.
			serviceAddress, err = c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Get(service.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// ServiceAddress definitely doesn't exist, so we can create it.
					return c.createNewServiceAddress(service)
				}
			}

			return nil, err
		}

		return nil, err
	}

	serviceAddress, err = c.syncExistingServiceAddress(service, serviceAddress)
	if err != nil {
		return nil, err
	}

	return serviceAddress, nil
}

func (c *Controller) syncExistingServiceAddress(
	service *latticev1.Service,
	address *latticev1.ServiceAddress,
) (*latticev1.ServiceAddress, error) {
	spec, err := c.serviceAddressSpec(service)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(address.Spec, spec) {
		return address, nil
	}

	glog.V(4).Infof("ServiceAddress %v for Service %v/%v had out of date spec, updating", address.Name, service.Namespace, service.Name)
	return c.updateServiceAddressSpec(address, spec)
}

func (c *Controller) updateServiceAddressSpec(
	address *latticev1.ServiceAddress,
	spec latticev1.ServiceAddressSpec,
) (*latticev1.ServiceAddress, error) {
	if reflect.DeepEqual(address.Spec, spec) {
		return address, nil
	}

	// Copy so the shared cache isn't mutated
	address = address.DeepCopy()
	address.Spec = spec

	return c.latticeClient.LatticeV1().ServiceAddresses(address.Namespace).Update(address)
}

func (c *Controller) createNewServiceAddress(service *latticev1.Service) (*latticev1.ServiceAddress, error) {
	serviceAddress, err := c.newServiceAddress(service)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Create(serviceAddress)
}

func (c *Controller) newServiceAddress(service *latticev1.Service) (*latticev1.ServiceAddress, error) {
	spec, err := c.serviceAddressSpec(service)
	if err != nil {
		return nil, err
	}

	serviceAddress := &latticev1.ServiceAddress{
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: spec,
		Status: latticev1.ServiceAddressStatus{
			State: latticev1.ServiceAddressStatePending,
		},
	}

	return serviceAddress, nil
}

func (c *Controller) serviceAddressSpec(service *latticev1.Service) (latticev1.ServiceAddressSpec, error) {
	endpointGroups := map[string]latticev1.ServiceAddressEndpointGroup{
		"service": {
			Service: &service.Name,
		},
	}

	ports := map[int32]latticev1.ServiceAddressPort{}
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			switch componentPort.Protocol {
			case block.ProtocolHTTP:
				httpPortConfig, err := c.serviceAddressHTTPPort(service, componentPort)
				if err != nil {
					return latticev1.ServiceAddressSpec{}, err
				}

				ports[componentPort.Port] = latticev1.ServiceAddressPort{
					HTTP: httpPortConfig,
				}

			case block.ProtocolTCP:
				tcpPortConfig, err := serviceAddressTCPPort(componentPort)
				if err != nil {
					return latticev1.ServiceAddressSpec{}, err
				}

				ports[componentPort.Port] = latticev1.ServiceAddressPort{
					TCP: tcpPortConfig,
				}
			}
		}
	}

	spec := latticev1.ServiceAddressSpec{
		Path:           service.Spec.Path,
		EndpointGroups: endpointGroups,
		Ports:          ports,
	}
	return spec, nil
}

func (c *Controller) serviceAddressHTTPPort(
	service *latticev1.Service,
	componentPort latticev1.ComponentPort,
) (*latticev1.ServiceAddressPortHTTPConfig, error) {
	serviceMeshPort, err := c.serviceMesh.ServiceMeshPort(service, componentPort.Port)
	if err != nil {
		return nil, err
	}

	target := latticev1.ServiceAddressPortHTTPTargetConfig{
		Port:          serviceMeshPort,
		EndpointGroup: "service",
		Weight:        100,
	}

	// FIXME(kevinrosendahl): add health check

	config := &latticev1.ServiceAddressPortHTTPConfig{
		Targets: []latticev1.ServiceAddressPortHTTPTargetConfig{target},
	}
	return config, nil
}

func serviceAddressTCPPort(componentPort latticev1.ComponentPort) (*latticev1.ServiceAddressPortTCPConfig, error) {
	config := &latticev1.ServiceAddressPortTCPConfig{
		EndpointGroup: "service",
	}

	// FIXME(kevinrosendahl): add health check
	return config, nil
}
