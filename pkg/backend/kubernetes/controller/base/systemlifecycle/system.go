package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	kubelabels "k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) getSystem(namespace string) (*crv1.System, error) {
	systems, err := c.systemLister.Systems(namespace).List(kubelabels.Everything())
	if err != nil {
		return nil, err
	}

	if len(systems) != 0 {
		return nil, fmt.Errorf("expected a System in namespace %v but found %v", namespace, len(systems))
	}

	return systems[0], nil
}

func (c *Controller) systemSpec(rollout *crv1.SystemRollout, build *crv1.SystemBuild) (*crv1.SystemSpec, error) {
	services, err := c.systemServices(rollout, build)
	if err != nil {
		return nil, err
	}

	sysSpec := &crv1.SystemSpec{
		Services: services,
	}

	return sysSpec, nil
}

func (c *Controller) systemServices(rollout *crv1.SystemRollout, build *crv1.SystemBuild) (map[tree.NodePath]crv1.SystemSpecServiceInfo, error) {
	if build.Status.State != crv1.SystemBuildStateSucceeded {
		return nil, fmt.Errorf("cannot get system services for build %v/%v, must be in state %v but is in %v", build.Namespace, build.Name, crv1.SystemBuildStateSucceeded, build.Status.State)
	}

	services := map[tree.NodePath]crv1.SystemSpecServiceInfo{}
	for path, service := range build.Spec.DefinitionRoot.Services() {
		serviceBuildInfo, ok := build.Spec.Services[path]
		if !ok {
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild does not have expected Service %v", path)
		}

		serviceBuild, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).Get(*serviceBuildInfo.Name)
		if err != nil {
			return nil, err
		}

		// Create crv1.ComponentBuildArtifacts for each Component in the Service
		componentBuildArtifacts := map[string]crv1.ComponentBuildArtifacts{}
		for component, componentBuildInfo := range serviceBuild.Spec.Components {
			if componentBuildInfo.Name == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("ServiceBuild %v Component %v does not have a ComponentBuildName", serviceBuild.Name, component)
			}

			componentBuild, err := c.componentBuildLister.ComponentBuilds(serviceBuild.Namespace).Get(*componentBuildInfo.Name)

			if err != nil {
				return nil, err
			}

			if componentBuild.Spec.Artifacts == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("ComponentBuild %v does not have Artifacts", *componentBuildInfo.Name)
			}

			componentBuildArtifacts[component] = *componentBuild.Spec.Artifacts
		}

		services[path] = crv1.SystemSpecServiceInfo{
			Definition:              *(service.Definition().(*definition.Service)),
			ComponentBuildArtifacts: componentBuildArtifacts,
		}
	}

	return services, nil
}
