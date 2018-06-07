package system

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/deckarep/golang-set"
	"github.com/satori/go.uuid"
)

func (c *Controller) syncSystemServices(system *latticev1.System) (map[tree.NodePath]latticev1.SystemStatusService, error) {
	// N.B.: as it currently is, this controller does not allow for a "move" i.e.
	// renaming a service (changing its path). When it comes time to implement that,
	// a possible approach would be to add an annotation indicating what moves need to be made,
	// then remove that annotation when updating the status. Will need to think through the idempotency.
	// Also when renaming a service it should probably also be done via an annotation, so other components
	// can continue to just look at the label as the confirmed path of the service, as opposed to trying
	// to figure out if a rename is in flight.
	services := make(map[tree.NodePath]latticev1.SystemStatusService)
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)
	serviceNames := mapset.NewSet()

	// Loop through the services defined in the system's Spec, and create/update any that need it
	for path, serviceInfo := range system.Spec.Services {
		var service *latticev1.Service

		serviceStatus, ok := system.Status.Services[path]
		if !ok {
			// If a status for this service path hasn't been set, then either we haven't created the service yet,
			// or we were unable to update the system's Status after creating the service

			// First check our cache to see if the service exists.
			var err error
			service, err = c.getServiceFromCache(systemNamespace, path)
			if err != nil {
				return nil, err
			}

			if service == nil {
				// The service wasn't in the cache, so do a quorum read to see if it was created.
				// N.B.: could first loop through and check to see if we need to do a quorum read
				// on any of the services, then just do one list.
				service, err = c.getServiceFromAPI(systemNamespace, path)
				if err != nil {
					return nil, err
				}

				if service == nil {
					// The service actually doesn't exist yet. Create it with a new UUID as the name.
					service, err = c.createNewService(system, &serviceInfo, path)
					if err != nil {
						return nil, err
					}

					// Successfully created the service. No need to check if it needs to be updated.
					services[path] = latticev1.SystemStatusService{
						Name:          service.Name,
						Generation:    service.Generation,
						ServiceStatus: service.Status,
					}
					serviceNames.Add(service.Name)
					continue
				}
			}
			// We were able to find an existing service for this path. We'll check below if it
			// needs to be updated.
		} else {
			// There is supposedly already a service for this path.
			serviceName := serviceStatus.Name
			var err error

			service, err = c.serviceLister.Services(systemNamespace).Get(serviceName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, fmt.Errorf("error trying to get cached service %v for %v", serviceName, system.Description())
				}

				// The service wasn't in the cache. Perhaps it was recently created. Do a quorum read.
				service, err = c.latticeClient.LatticeV1().Services(systemNamespace).Get(serviceName, metav1.GetOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						return nil, fmt.Errorf("error trying to get service %v for %v", serviceName, system.Description())
					}

					// FIXME: should we just recreate the service here?
					// what happens when a deploy doesnt fully succeed and there's a leftover terminating service with
					// the same path as a new service?
					return nil, fmt.Errorf("%v has reference to non existant service %v", system.Description(), serviceName)
				}
			}
		}

		// We found an existing service. Calculate what its Spec should look like,
		// and update the service if its current Spec is different.
		spec, err := c.serviceSpec(system, &serviceInfo, path)
		if err != nil {
			return nil, fmt.Errorf("error getting desired spec for service %v in %v: %v", path.String(), system.Description(), err)
		}

		service, err = c.updateService(service, spec, path)
		if err != nil {
			return nil, err
		}

		serviceNames.Add(service.Name)
		services[path] = latticev1.SystemStatusService{
			Name:          service.Name,
			Generation:    service.Generation,
			ServiceStatus: service.Status,
		}
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO(kevinrosendahl): should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	allServices, err := c.serviceLister.Services(systemNamespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, service := range allServices {
		if !serviceNames.Contains(service.Name) {
			if service.DeletionTimestamp == nil {
				err := c.deleteService(service)
				if err != nil {
					return nil, err
				}
			}

			path, err := service.PathLabel()
			if err != nil {
				// FIXME: warn
				continue
			}

			// copy so the shared cache isn't mutated
			status := service.Status.DeepCopy()
			status.State = latticev1.ServiceStateDeleting

			services[path] = latticev1.SystemStatusService{
				Name:          service.Name,
				Generation:    service.Generation,
				ServiceStatus: *status,
			}
		}
	}

	return services, nil
}

func (c *Controller) createNewService(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (*latticev1.Service, error) {
	service, err := c.newService(system, serviceInfo, path)
	if err != nil {
		return nil, fmt.Errorf("error getting new service for %v in %v: %v", path.String(), system.Description(), err)
	}

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).Create(service)
	if err != nil {
		return nil, fmt.Errorf("error creating new service for %v in %v: %v", path.String(), system.Description(), err)
	}

	return result, nil
}

func (c *Controller) newService(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (*latticev1.Service, error) {
	spec, err := c.serviceSpec(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	systemNamespace := system.ResourceNamespace(c.namespacePrefix)

	service := &latticev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       systemNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, latticev1.SystemKind)},
			Labels: map[string]string{
				latticev1.ServicePathLabelKey: path.ToDomain(),
			},
		},
		Spec: spec,
	}

	annotations, err := c.serviceMesh.ServiceAnnotations(service)
	if err != nil {
		return nil, err
	}

	service.Annotations = annotations

	return service, nil
}

func (c *Controller) serviceSpec(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (latticev1.ServiceSpec, error) {
	var numInstances int32
	if serviceInfo.Definition.Resources.NumInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources.NumInstances)
	} else if serviceInfo.Definition.Resources.MinInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources.MinInstances)
	} else {
		err := fmt.Errorf(
			"service %v (%v) invalid definition: num_instances or min_instances must be set",
			path.ToDomain(),
			system.V1ID(),
		)
		return latticev1.ServiceSpec{}, err
	}

	componentPorts := map[string][]latticev1.ContainerPort{}

	for _, component := range serviceInfo.Definition.Components {
		var ports []latticev1.ContainerPort
		for _, port := range component.Ports {
			componentPort := latticev1.ContainerPort{
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

	spec := latticev1.ServiceSpec{
		Definition:              serviceInfo.Definition,
		ContainerBuildArtifacts: serviceInfo.ContainerBuildArtifacts,
		Ports:        componentPorts,
		NumInstances: numInstances,
	}
	return spec, nil
}

func (c *Controller) deleteService(service *latticev1.Service) error {
	// background delete will add deletionTimestamp to the service, but will not
	// try to act upon any of the dependents since the service has a finalizer
	// this allows us to clean up the service in a controlled way
	backgroundDelete := metav1.DeletePropagationBackground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &backgroundDelete,
	}

	err := c.latticeClient.LatticeV1().Services(service.Namespace).Delete(service.Name, deleteOptions)
	if err != nil {
		return fmt.Errorf("error deleting %v: %v", service.Description(c.namespacePrefix), err)
	}

	return nil
}

func (c *Controller) updateService(service *latticev1.Service, spec latticev1.ServiceSpec, path tree.NodePath) (*latticev1.Service, error) {
	if !c.serviceNeedsUpdate(service, spec, path) {
		return service, nil
	}

	// Copy so the cache isn't mutated
	service = service.DeepCopy()
	service.Spec = spec

	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}
	service.Labels[latticev1.ServicePathLabelKey] = path.ToDomain()

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
	if err != nil {
		return nil, fmt.Errorf("error updating %v: %v", service.Description(c.namespacePrefix), err)
	}

	return result, err
}

func (c *Controller) serviceNeedsUpdate(service *latticev1.Service, spec latticev1.ServiceSpec, path tree.NodePath) bool {
	if !reflect.DeepEqual(service.Spec, spec) {
		return true
	}

	currentPath, err := service.PathLabel()
	if err != nil {
		return true
	}

	return currentPath != path
}

func (c *Controller) getServiceFromCache(namespace string, path tree.NodePath) (*latticev1.Service, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, fmt.Errorf("error getting selector for cached service %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	services, err := c.serviceLister.Services(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("error getting cached services in namespace %v", namespace)
	}

	if len(services) == 0 {
		return nil, nil
	}

	if len(services) > 1 {
		return nil, fmt.Errorf("found multiple cached services with path %v in namespace %v", path.String(), namespace)
	}

	return services[0], nil
}

func (c *Controller) getServiceFromAPI(namespace string, path tree.NodePath) (*latticev1.Service, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, fmt.Errorf("error getting selector for  service %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	services, err := c.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("error getting services in namespace %v", namespace)
	}

	if len(services.Items) == 0 {
		return nil, nil
	}

	if len(services.Items) > 1 {
		return nil, fmt.Errorf("found multiple services with path %v in namespace %v", path.String(), namespace)
	}

	return &services.Items[0], nil
}
