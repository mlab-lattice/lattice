package system

import (
	"fmt"
	"reflect"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncSystemServices(system *crv1.System) (map[tree.NodePath]string, map[string]crv1.ServiceStatus, []string, error) {
	// Maps Service path to Service.Name of the Service
	services := map[tree.NodePath]string{}

	// Maps Service.Name to Service.Status
	serviceStatuses := map[string]crv1.ServiceStatus{}

	// Loop through the Services defined in the System's Spec, and create/update any that need it
	for path, serviceInfo := range system.Spec.Services {
		var service *crv1.Service

		serviceName, ok := system.Status.Services[path]
		if !ok {
			pathDomain := path.ToDomain(true)
			// We don't have the name of the Service in our Status, but it may still have been created already.
			// First, look in the cache for a Service with the proper label.
			selector := kubelabels.NewSelector()
			requirement, err := kubelabels.NewRequirement(kubeconstants.LabelKeyServicePathDomain, selection.Equals, []string{pathDomain})
			if err != nil {
				return nil, nil, nil, err
			}

			selector = selector.Add(*requirement)
			services, err := c.serviceLister.Services(system.Namespace).List(selector)
			if err != nil {
				return nil, nil, nil, err
			}

			if len(services) > 1 {
				return nil, nil, nil, fmt.Errorf("multiple Services in the %v namespace are labeled with %v = %v", system.Namespace, kubeconstants.LabelKeyServicePathDomain, pathDomain)
			}

			if len(services) == 1 {
				service = services[0]
			}

			if len(services) == 0 {
				// The cache did not have a Service matching the label.
				// However, it would be a constraint violation to have multiple Services for the same path,
				// so we'll have to do a quorum read from the API to make sure that the Service does not exist.
				services, err := c.latticeClient.LatticeV1().Services(system.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
				if err != nil {
					return nil, nil, nil, err
				}

				if len(services.Items) > 1 {
					return nil, nil, nil, fmt.Errorf("multiple Services in the %v namespace are labeled with %v = %v", system.Namespace, kubeconstants.LabelKeyServicePathDomain, pathDomain)
				}

				if len(services.Items) == 1 {
					service = &services.Items[0]
				}

				if len(services.Items) == 0 {
					// We are now sure that the Service does not exist, so now we can create it.
					service, err = c.createNewService(system, &serviceInfo, path)
				}
			}
		}

		if service == nil {
			var err error
			service, err = c.serviceLister.Services(system.Namespace).Get(serviceName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, nil, nil, err
				}

				// the Service wasn't in our cache, so check with the API
				service, err = c.latticeClient.LatticeV1().Services(system.Namespace).Get(serviceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						// FIXME: send warn event
						// TODO: should we just create a new Service here?
						return nil, nil, nil, fmt.Errorf(
							"Service %v in namespace %v has Name %v but Service does not exist",
							path,
							system.Namespace,
							serviceName,
						)
					}

					return nil, nil, nil, err
				}
			}
		}

		// Otherwise, get a new spec and update the service
		spec, err := serviceSpec(system, &serviceInfo, path)
		if err != nil {
			return nil, nil, nil, err
		}

		service, err = c.updateService(service, spec)
		if err != nil {
			return nil, nil, nil, err
		}

		services[path] = service.Name
		serviceStatuses[service.Name] = service.Status
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO(kevinrosendahl): should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	allServices, err := c.serviceLister.Services(system.Namespace).List(kubelabels.Everything())
	if err != nil {
		return nil, nil, nil, err
	}

	var deletedServices []string
	for _, service := range allServices {
		if _, ok := serviceStatuses[service.Name]; !ok {
			glog.V(4).Infof("Found Service %q in Namespace %q that is no longer in the System Spec", service.Name, service.Namespace)
			deletedServices = append(deletedServices, service.Name)

			if service.DeletionTimestamp == nil {
				err := c.latticeClient.LatticeV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})
				if err != nil {
					return nil, nil, nil, err
				}
			}
		}
	}

	return services, serviceStatuses, deletedServices, nil
}

func (c *Controller) createNewService(system *crv1.System, serviceInfo *crv1.SystemSpecServiceInfo, path tree.NodePath) (*crv1.Service, error) {
	service, err := c.newService(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().Services(service.Namespace).Create(service)
}

func (c *Controller) newService(system *crv1.System, serviceInfo *crv1.SystemSpecServiceInfo, path tree.NodePath) (*crv1.Service, error) {
	labels := map[string]string{
		kubeconstants.LabelKeyServicePathDomain: path.ToDomain(true),
	}

	spec, err := serviceSpec(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	service := &crv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       system.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, controllerKind)},
			Labels:          labels,
		},
		Spec: spec,
		Status: crv1.ServiceStatus{
			State: crv1.ServiceStatePending,
		},
	}

	annotations, err := c.serviceMesh.ServiceAnnotations(service)
	if err != nil {
		return nil, err
	}

	service.Annotations = annotations

	return service, nil
}

func serviceSpec(system *crv1.System, serviceInfo *crv1.SystemSpecServiceInfo, path tree.NodePath) (crv1.ServiceSpec, error) {
	var numInstances int32
	if serviceInfo.Definition.Resources().NumInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().NumInstances)
	} else if serviceInfo.Definition.Resources().MinInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().MinInstances)
	} else {
		err := fmt.Errorf(
			"System %v/%v Service %v invalid Service definition: num_instances or min_instances must be set",
			system.Namespace,
			system.Name,
			path,
		)
		return crv1.ServiceSpec{}, err
	}

	componentPorts := map[string][]crv1.ComponentPort{}

	for _, component := range serviceInfo.Definition.Components() {
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
		}

		componentPorts[component.Name] = ports
	}

	spec := crv1.ServiceSpec{
		Path:                    path,
		Definition:              serviceInfo.Definition,
		ComponentBuildArtifacts: serviceInfo.ComponentBuildArtifacts,
		Ports:        componentPorts,
		NumInstances: numInstances,
	}
	return spec, nil
}

func (c *Controller) updateService(service *crv1.Service, spec crv1.ServiceSpec) (*crv1.Service, error) {
	if reflect.DeepEqual(service.Spec, spec) {
		return service, nil
	}

	// Copy so the cache isn't mutated
	service = service.DeepCopy()
	service.Spec = spec

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}
