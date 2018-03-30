package system

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
)

func (c *Controller) syncSystemServices(system *latticev1.System) (map[tree.NodePath]latticev1.ServiceStatus, []tree.NodePath, error) {
	services := make(map[tree.NodePath]latticev1.ServiceStatus)
	systemNamespace := kubeutil.SystemNamespace(c.latticeID, system.V1ID())

	// Loop through the services defined in the system's Spec, and create/update any that need it
	for path, serviceInfo := range system.Spec.Services {
		var service *latticev1.Service
		pathDomain := path.ToDomain()

		// First, look in the cache to see if the Service already exists
		var err error
		service, err = c.serviceLister.Services(systemNamespace).Get(pathDomain)
		if err != nil {
			if !errors.IsNotFound(err) {
				return nil, nil, err
			}

			// If the service didn't exist in our cache, try to create it
			service, err = c.createNewService(system, &serviceInfo, path)
			if err == nil {
				// If we created the service, no need to do any more work on it.
				services[path] = service.Status
				continue
			}

			// There was some unexpected error creating the service.
			if !errors.IsAlreadyExists(err) {
				return nil, nil, err
			}

			// The service didn't exist in our cache, but it does exist in the server.
			// Retrieve it with a quorum read.
			service, err = c.latticeClient.LatticeV1().Services(systemNamespace).Get(pathDomain, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, nil, err
				}

				return nil, nil, fmt.Errorf("could not create service %v (%v) because it already existed, but it does not exist", pathDomain, system.Name)
			}
		}

		// We found an existing service. Calculate what its Spec should look like,
		// and update the service if its current Spec is different.
		spec, err := c.serviceSpec(system, &serviceInfo, path)
		if err != nil {
			return nil, nil, err
		}

		service, err = c.updateService(service, spec)
		if err != nil {
			return nil, nil, err
		}

		services[path] = service.Status
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO(kevinrosendahl): should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	allServices, err := c.serviceLister.Services(systemNamespace).List(kubelabels.Everything())
	if err != nil {
		return nil, nil, err
	}

	var deletedServices []tree.NodePath
	for _, service := range allServices {
		servicePath, err := tree.NodePathFromDomain(service.Name)
		if err != nil {
			return nil, nil, err
		}

		if _, ok := services[servicePath]; !ok {
			glog.V(4).Infof(
				"Found service %v (%v) that is no longer in the system Spec",
				service.Name,
				system.Name,
			)
			deletedServices = append(deletedServices, servicePath)

			if service.DeletionTimestamp == nil {
				err := c.latticeClient.LatticeV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}

	return services, deletedServices, nil
}

func (c *Controller) createNewService(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (*latticev1.Service, error) {
	service, err := c.newService(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().Services(service.Namespace).Create(service)
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

	systemNamespace := kubeutil.SystemNamespace(c.latticeID, system.V1ID())
	pathDomain := path.ToDomain()

	service := &latticev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pathDomain,
			Namespace:       systemNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, controllerKind)},
		},
		Spec: spec,
		Status: latticev1.ServiceStatus{
			State: latticev1.ServiceStatePending,
		},
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
	if serviceInfo.Definition.Resources().NumInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().NumInstances)
	} else if serviceInfo.Definition.Resources().MinInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().MinInstances)
	} else {
		err := fmt.Errorf(
			"service %v (%v) invalid definition: num_instances or min_instances must be set",
			path.ToDomain(),
			system.Name,
		)
		return latticev1.ServiceSpec{}, err
	}

	componentPorts := map[string][]latticev1.ComponentPort{}

	for _, component := range serviceInfo.Definition.Components() {
		var ports []latticev1.ComponentPort
		for _, port := range component.Ports {
			componentPort := latticev1.ComponentPort{
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
		ComponentBuildArtifacts: serviceInfo.ComponentBuildArtifacts,
		Ports:        componentPorts,
		NumInstances: numInstances,
	}
	return spec, nil
}

func (c *Controller) updateService(service *latticev1.Service, spec latticev1.ServiceSpec) (*latticev1.Service, error) {
	if reflect.DeepEqual(service.Spec, spec) {
		return service, nil
	}

	// Copy so the cache isn't mutated
	service = service.DeepCopy()
	service.Spec = spec

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}
