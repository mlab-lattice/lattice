package service

import (
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/block"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncServiceServiceAddress(service *crv1.Service) (*crv1.ServiceAddress, error) {
	serviceAddress, err := c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Get(service.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createNewServiceAddress(service)
		}

		return nil, err
	}

	serviceAddress, err = c.syncExistingServiceAddress(service, serviceAddress)
	if err != nil {
		return nil, err
	}

	return serviceAddress, nil
}

func (c *Controller) syncExistingServiceAddress(service *crv1.Service, serviceAddress *crv1.ServiceAddress) (*crv1.ServiceAddress, error) {
	desiredSpec, err := serviceAddressSpec(service)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(serviceAddress.Spec, desiredSpec) {
		return serviceAddress, nil
	}

	glog.V(4).Infof("ServiceAddress for Service %v/%v had out of date spec, updating", serviceAddress.Name, service.Namespace, service.Name)
	serviceAddress.Spec = *desiredSpec
	return c.latticeClient.LatticeV1().ServiceAddresses(serviceAddress.Namespace).Update(serviceAddress)
}

func (c *Controller) createNewServiceAddress(service *crv1.Service) (*crv1.ServiceAddress, error) {
	serviceAddress, err := newServiceAddress(service)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Create(serviceAddress)
}

func newServiceAddress(service *crv1.Service) (*crv1.ServiceAddress, error) {
	spec, err := serviceAddressSpec(service)
	if err != nil {
		return nil, err
	}

	serviceAddress := &crv1.ServiceAddress{
		ObjectMeta: metav1.ObjectMeta{
			Name: service.Name,
		},
		Spec: *spec,
		Status: crv1.ServiceAddressStatus{
			State: crv1.ServiceAddressStatePending,
		},
	}

	return serviceAddress, nil
}

func serviceAddressSpec(service *crv1.Service) (*crv1.ServiceAddressSpec, error) {
	endpointGroups := map[string]crv1.ServiceAddressEndpointGroup{
		"service": {
			Service: &service.Name,
		},
	}

	ports := map[int32]crv1.ServiceAddressPort{}
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			switch componentPort.Protocol {
			case block.ProtocolHTTP:
				httpPortConfig, err := serviceAddressHTTPPort(componentPort)
				if err != nil {
					return nil, err
				}

				ports[componentPort.Port] = crv1.ServiceAddressPort{
					HTTP: httpPortConfig,
				}

			case block.ProtocolTCP:
				tcpPortConfig, err := serviceAddressTCPPort(componentPort)
				if err != nil {
					return nil, err
				}

				ports[componentPort.Port] = crv1.ServiceAddressPort{
					TCP: tcpPortConfig,
				}
			}
		}
	}

	spec := &crv1.ServiceAddressSpec{
		EndpointGroups: endpointGroups,
		Ports:          ports,
	}
	return spec, nil
}

func serviceAddressHTTPPort(componentPort crv1.ComponentPort) (*crv1.ServiceAddressPortHTTPConfig, error) {
	target := crv1.ServiceAddressPortHTTPTargetConfig{
		Port:          componentPort.EnvoyPort,
		EndpointGroup: "service",
		Weight:        100,
	}

	// FIXME(kevinrosendahl): add health check

	config := &crv1.ServiceAddressPortHTTPConfig{
		Targets: []crv1.ServiceAddressPortHTTPTargetConfig{target},
	}
	return config, nil
}

func serviceAddressTCPPort(componentPort crv1.ComponentPort) (*crv1.ServiceAddressPortTCPConfig, error) {
	config := &crv1.ServiceAddressPortTCPConfig{
		EndpointGroup: "service",
	}

	// FIXME(kevinrosendahl): add health check
	return config, nil
}
