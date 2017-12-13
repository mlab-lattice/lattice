package system

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

func (c *Controller) syncSystemServices(system *crv1.System) (*crv1.System, error) {
	validServiceNames := map[string]bool{}

	// Loop through the Services defined in the System's Spec, and create/update any that need it
	servicesInfo := map[tree.NodePath]crv1.SystemServicesInfo{}
	for path, serviceInfo := range system.Spec.Services {
		// If the Service doesn't exist already, create one.
		if serviceInfo.Name == nil {
			glog.V(5).Infof("Did not find a Service for %q, creating one", path)
			service, err := c.createNewService(system, &serviceInfo, path)
			if err != nil {
				return nil, err
			}
			serviceInfo.Name = &service.Name
			serviceInfo.State = &service.Status.State

			servicesInfo[path] = serviceInfo
			validServiceNames[service.Name] = true
			continue
		}

		// A Service has already been created. Check if its definition is the same
		// definition. We'll assume that the rest of the spec is properly formed.
		service, err := c.serviceLister.Services(system.Namespace).Get(*serviceInfo.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				// FIXME: send warn event
				// TODO: should we just create a new Service here?
				return nil, fmt.Errorf(
					"Service %v has Name %v but Service does not exist",
					path,
					serviceInfo.Name,
				)
			}

			return nil, err
		}

		validServiceNames[service.Name] = true

		// If the definitions are the same, there's nothing to update.
		// FIXME(kevinrosendahl): should we get a new spec here? If calculating envoy ports is cheap *and* deterministic, we probably should.
		// If we don't, and we update what a Service.Spec looks like for a given definition, it may never get updated.
		if reflect.DeepEqual(serviceInfo.Definition, service.Spec.Definition) {
			servicesInfo[path] = serviceInfo
			continue
		}

		// Otherwise, get a new spec and update the service
		spec, err := serviceSpec(&serviceInfo, path)
		if err != nil {
			return nil, err
		}

		service.Spec = *spec
		service.Status.State = crv1.ServiceStateUpdating

		service, err = c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
		if err != nil {
			return nil, err
		}

		serviceInfo.State = &service.Status.State
		servicesInfo[path] = serviceInfo
	}

	// update system if necessary
	if !reflect.DeepEqual(system.Spec.Services, servicesInfo) {
		// Copy system so the shared cache isn't mutated
		system = system.DeepCopy()
		system.Spec.Services = servicesInfo

		var err error
		system, err = c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
		if err != nil {
			return nil, err
		}
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO(kevinrosendahl): should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	services, err := c.serviceLister.Services(system.Namespace).List(kubelabels.Everything())
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if _, ok := validServiceNames[service.Name]; !ok {
			glog.V(4).Infof("Found Service %q in Namespace %q that is no longer in the System Spec", service.Name, service.Namespace)
			err := c.latticeClient.LatticeV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})
			if err != nil {
				return nil, err
			}
		}
	}

	return system, nil
}

func (c *Controller) getService(namespace, name string) (*crv1.Service, error) {
	svc, err := c.serviceLister.Services(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return svc, nil
}

func (c *Controller) createNewService(system *crv1.System, serviceInfo *crv1.SystemServicesInfo, path tree.NodePath) (*crv1.Service, error) {
	svc, err := newService(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().Services(svc.Namespace).Create(svc)
}

func newService(system *crv1.System, serviceInfo *crv1.SystemServicesInfo, path tree.NodePath) (*crv1.Service, error) {
	spec, err := serviceSpec(serviceInfo, path)
	if err != nil {
		return nil, err
	}

	service := &crv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       system.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, controllerKind)},
		},
		Spec: *spec,
		Status: crv1.ServiceStatus{
			State: crv1.ServiceStatePending,
		},
	}
	return service, nil
}

func serviceSpec(serviceInfo *crv1.SystemServicesInfo, path tree.NodePath) (*crv1.ServiceSpec, error) {
	componentPorts, portSet := servicePorts(serviceInfo)
	envoyPorts, err := envoyPorts(portSet)
	if err != nil {
		return nil, err
	}

	componentPorts, remainingEnvoyPorts, err := assignEnvoyPorts(serviceInfo.Definition.Components, componentPorts, envoyPorts)
	if err != nil {
		return nil, err
	}

	if len(remainingEnvoyPorts) != 2 {
		return nil, fmt.Errorf("expected 2 remaining envoy ports, got %v", len(remainingEnvoyPorts))
	}

	envoyAdminPort := remainingEnvoyPorts[0]
	envoyEgressPort := remainingEnvoyPorts[1]

	spec := &crv1.ServiceSpec{
		Path:       path,
		Definition: serviceInfo.Definition,

		ComponentBuildArtifacts: serviceInfo.ComponentBuildArtifacts,

		Ports:           componentPorts,
		EnvoyAdminPort:  envoyAdminPort,
		EnvoyEgressPort: envoyEgressPort,
	}
	return spec, nil
}

func servicePorts(serviceInfo *crv1.SystemServicesInfo) (map[string][]crv1.ComponentPort, map[int32]struct{}) {
	componentPorts := map[string][]crv1.ComponentPort{}
	portSet := map[int32]struct{}{}

	for _, component := range serviceInfo.Definition.Components {
		var ports []crv1.ComponentPort
		for _, port := range component.Ports {
			componentPort := crv1.ComponentPort{
				Name:     port.Name,
				Port:     int32(port.Port),
				Protocol: port.Protocol,
				Public:   false,
			}

			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				componentPort.Public = true
			}

			ports = append(ports, componentPort)
			portSet[int32(port.Port)] = struct{}{}
		}

		componentPorts[component.Name] = ports
	}

	return componentPorts, portSet
}

func envoyPorts(portSet map[int32]struct{}) ([]int32, error) {
	var envoyPortIdx int32 = 10000
	var envoyPorts []int32

	// Need to find len(portSet) + 2 unique ports to use for envoy
	// (one for egress, one for admin, and one per component port for ingress)
	for i := 0; i <= len(portSet)+1; i++ {

		// Loop up to len(portSet) + 1 times to find an unused port
		// we can use for envoy.
		for j := 0; j <= len(portSet); j++ {

			// If the current envoyPortIdx is not being used by a component,
			// we'll use it for envoy. Otherwise, on to the next one.
			currPortIdx := envoyPortIdx
			envoyPortIdx++

			if _, ok := portSet[currPortIdx]; !ok {
				envoyPorts = append(envoyPorts, currPortIdx)
				break
			}
		}
	}

	if len(envoyPorts) != len(portSet)+2 {
		return nil, fmt.Errorf("expected %v envoy ports but got %v", len(portSet)+1, len(envoyPorts))
	}

	return envoyPorts, nil
}

func assignEnvoyPorts(components []*block.Component, componentPorts map[string][]crv1.ComponentPort, envoyPorts []int32) (map[string][]crv1.ComponentPort, []int32, error) {
	// Assign an envoy port to each component port, and pop the used envoy port off the slice each time.
	for _, component := range components {
		var ports []crv1.ComponentPort

		for _, componentPort := range componentPorts[component.Name] {
			if len(envoyPorts) == 0 {
				return nil, nil, fmt.Errorf("ran out of ports when assigning envoyPorts")
			}

			componentPort.EnvoyPort = envoyPorts[0]
			ports = append(ports, componentPort)
			envoyPorts = envoyPorts[1:]
		}

		componentPorts[component.Name] = ports
	}

	return componentPorts, envoyPorts, nil
}
