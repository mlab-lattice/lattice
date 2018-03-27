package systemlifecycle

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) updateSystem(
	system *latticev1.System,
	services map[tree.NodePath]latticev1.SystemSpecServiceInfo,
) (*latticev1.System, error) {
	spec := system.Spec.DeepCopy()
	spec.Services = services

	return c.updateSystemSpec(system, *spec)
}

func (c *Controller) updateSystemSpec(system *latticev1.System, spec latticev1.SystemSpec) (*latticev1.System, error) {
	if reflect.DeepEqual(system.Spec, spec) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Spec = spec

	// FIXME: remove this when ObservedGeneration is supported for CRD
	system.Status.UpdateProcessed = false

	return c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
}

func isSystemStatusCurrent(system *latticev1.System) bool {
	return system.Status.UpdateProcessed
	// FIXME: go back to this when ObservedGeneration is supported for CRD
	//return system.Status.ObservedGeneration == system.Generation
}

func (c *Controller) getSystem(namespace string) (*latticev1.System, error) {
	systemID, err := kubeutil.SystemID(namespace)
	if err != nil {
		return nil, err
	}

	internalNamespace := kubeutil.InternalNamespace(c.latticeID)
	return c.systemLister.Systems(internalNamespace).Get(string(systemID))
}

func (c *Controller) systemSpec(rollout *latticev1.Deploy, build *latticev1.Build) (latticev1.SystemSpec, error) {
	services, err := c.systemServices(rollout, build)
	if err != nil {
		return latticev1.SystemSpec{}, err
	}

	spec := latticev1.SystemSpec{
		Services: services,
	}

	return spec, nil
}

func (c *Controller) systemServices(
	rollout *latticev1.Deploy,
	build *latticev1.Build,
) (map[tree.NodePath]latticev1.SystemSpecServiceInfo, error) {
	if build.Status.State != latticev1.BuildStateSucceeded {
		err := fmt.Errorf(
			"cannot get system services for build %v/%v, must be in state %v but is in %v",
			build.Namespace,
			build.Name,
			latticev1.BuildStateSucceeded,
			build.Status.State,
		)
		return nil, err
	}

	services := map[tree.NodePath]latticev1.SystemSpecServiceInfo{}
	for path, service := range build.Spec.DefinitionRoot.Services() {
		serviceBuildName, ok := build.Status.ServiceBuilds[path]
		if !ok {
			// FIXME: send warn event
			err := fmt.Errorf("Build %v/%v does not have expected Service %v", build.Namespace, build.Name, path)
			return nil, err
		}

		serviceBuild, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).Get(serviceBuildName)
		if err != nil {
			if errors.IsNotFound(err) {
				err = fmt.Errorf(
					"Build %v/%v has ServiceBuild %v for Service %v but it does not exist",
					build.Namespace,
					build.Name,
					serviceBuildName,
					path,
				)
				return nil, err
			}
			return nil, err
		}

		// Create latticev1.ComponentBuildArtifacts for each Component in the Service
		componentBuildArtifacts := map[string]latticev1.ComponentBuildArtifacts{}
		for component := range serviceBuild.Spec.Components {
			componentBuildName, ok := serviceBuild.Status.ComponentBuilds[component]
			if !ok {
				err := fmt.Errorf(
					"ServiceBuild %v/%v component %v does not have a ComponentBuild",
					serviceBuild.Namespace,
					serviceBuild.Name,
					component,
				)
				return nil, err
			}

			componentBuildStatus, ok := serviceBuild.Status.ComponentBuildStatuses[componentBuildName]
			if !ok {
				err := fmt.Errorf(
					"ServiceBuild %v/%v ComponentBuild %v does not have a ComponentBuildStatus",
					serviceBuild.Namespace,
					serviceBuild.Name,
					componentBuildName,
				)
				return nil, err
			}

			if componentBuildStatus.Artifacts == nil {
				// FIXME: send warn event
				err := fmt.Errorf("ComponentBuild %v/%v Status does not have Artifacts", build.Namespace, componentBuildName)
				return nil, err
			}

			componentBuildArtifacts[component] = *componentBuildStatus.Artifacts
		}

		services[path] = latticev1.SystemSpecServiceInfo{
			Definition:              service.Definition().(definition.Service),
			ComponentBuildArtifacts: componentBuildArtifacts,
		}
	}

	return services, nil
}
