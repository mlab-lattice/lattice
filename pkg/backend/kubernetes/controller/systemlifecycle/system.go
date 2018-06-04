package systemlifecycle

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) updateSystemLabels(
	system *latticev1.System,
	version *v1.SystemVersion,
	deployID *v1.DeployID,
	buildID *v1.BuildID,
) (*latticev1.System, error) {
	labels := make(map[string]string)
	for k, v := range system.Labels {
		labels[k] = v
	}

	delete(labels, latticev1.SystemDefinitionVersionLabelKey)
	if version != nil {
		labels[latticev1.SystemDefinitionVersionLabelKey] = string(*version)
	}

	delete(labels, latticev1.DeployIDLabelKey)
	if deployID != nil {
		labels[latticev1.DeployIDLabelKey] = string(*deployID)
	}

	delete(labels, latticev1.BuildIDLabelKey)
	if buildID != nil {
		labels[latticev1.BuildIDLabelKey] = string(*buildID)
	}

	if reflect.DeepEqual(labels, system.Labels) {
		return system, nil
	}

	// Copy so the original object isn't mutated
	system = system.DeepCopy()
	system.Labels = labels

	result, err := c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
	if err != nil {
		return nil, fmt.Errorf("error updating labels for %v: %v", system.Description(), err)
	}

	return result, nil
}

func (c *Controller) updateSystem(
	system *latticev1.System,
	services map[tree.NodePath]latticev1.SystemSpecServiceInfo,
	nodePools map[string]latticev1.NodePoolSpec,
) (*latticev1.System, error) {
	spec := system.Spec.DeepCopy()
	spec.Services = services
	spec.NodePools = nodePools

	return c.updateSystemSpec(system, *spec)
}

func (c *Controller) updateSystemSpec(system *latticev1.System, spec latticev1.SystemSpec) (*latticev1.System, error) {
	if reflect.DeepEqual(system.Spec, spec) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Spec = spec

	result, err := c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
	if err != nil {
		return nil, fmt.Errorf("error updating %v spec: %v", system.Description(), err)
	}

	return result, nil
}

func (c *Controller) getSystem(namespace string) (*latticev1.System, error) {
	systemID, err := kubeutil.SystemID(c.namespacePrefix, namespace)
	if err != nil {
		return nil, err
	}

	internalNamespace := kubeutil.InternalNamespace(c.namespacePrefix)
	result, err := c.systemLister.Systems(internalNamespace).Get(string(systemID))
	if err != nil {
		return nil, fmt.Errorf("error getting system for namespace %v: %v", namespace, err)
	}

	return result, nil
}

func (c *Controller) systemServices(
	build *latticev1.Build,
) (map[tree.NodePath]latticev1.SystemSpecServiceInfo, error) {
	if build.Status.State != latticev1.BuildStateSucceeded {
		err := fmt.Errorf(
			"cannot get services for %v, must be in state %v but is in %v",
			build.Description(c.namespacePrefix),
			latticev1.BuildStateSucceeded,
			build.Status.State,
		)
		return nil, err
	}

	services := make(map[tree.NodePath]latticev1.SystemSpecServiceInfo)
	for path, service := range build.Spec.Definition.Services() {
		serviceBuildName, ok := build.Status.ServiceBuilds[path]
		if !ok {
			// FIXME: send warn event
			err := fmt.Errorf(
				"%v does not have expected serviced %v",
				build.Description(c.namespacePrefix),
				path.String(),
			)
			return nil, err
		}

		serviceBuild, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).Get(serviceBuildName)
		if err != nil {
			if errors.IsNotFound(err) {
				err = fmt.Errorf(
					"%v has service build %v for service %v but it does not exist",
					build.Description(c.namespacePrefix),
					serviceBuildName,
					path,
				)
				return nil, err
			}
			return nil, err
		}

		// Create ComponentBuildArtifacts for each Component in the Service
		componentBuildArtifacts := make(map[string]latticev1.ComponentBuildArtifacts)
		for component := range serviceBuild.Spec.Components {
			componentBuildName, ok := serviceBuild.Status.ComponentBuilds[component]
			if !ok {
				err := fmt.Errorf(
					"%v component %v does not have a component build",
					serviceBuild.Description(c.namespacePrefix),
					component,
				)
				return nil, err
			}

			componentBuildStatus, ok := serviceBuild.Status.ComponentBuildStatuses[componentBuildName]
			if !ok {
				err := fmt.Errorf(
					"%v component build %v does not have a status",
					serviceBuild.Description(c.namespacePrefix),
					componentBuildName,
				)
				return nil, err
			}

			if componentBuildStatus.Artifacts == nil {
				// FIXME: send warn event
				err := fmt.Errorf(
					"%v component build %v status does not have artifacts",
					build.Description(c.namespacePrefix),
					componentBuildName,
				)
				return nil, err
			}

			componentBuildArtifacts[component] = *componentBuildStatus.Artifacts
		}

		services[path] = latticev1.SystemSpecServiceInfo{
			Definition:              service.Definition().(*definition.Service),
			ComponentBuildArtifacts: componentBuildArtifacts,
		}
	}

	return services, nil
}

func (c *Controller) systemNodePools(
	build *latticev1.Build,
) (map[string]latticev1.NodePoolSpec, error) {
	nodePools := make(map[string]latticev1.NodePoolSpec)
	err := tree.Walk(build.Spec.Definition, func(n tree.Node) error {
		system, ok := n.(*tree.SystemNode)
		if !ok {
			return nil
		}

		path := system.Path()
		pools := system.NodePools()

		for name, nodePool := range pools {
			p := v1.NewSystemSharedNodePoolPath(path, name)
			spec := latticev1.NodePoolSpec{
				NumInstances: nodePool.NumInstances,
				InstanceType: nodePool.InstanceType,
			}
			nodePools[p.String()] = spec
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return nodePools, nil
}
