package systemlifecycle

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
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
	definition *resolver.ComponentTree,
	artifacts *latticev1.SystemSpecWorkloadBuildArtifacts,
) (*latticev1.System, error) {
	spec := system.Spec.DeepCopy()
	spec.Definition = definition
	spec.WorkloadBuildArtifacts = artifacts

	return c.updateSystemSpec(system, spec)
}

func (c *Controller) updateSystemSpec(system *latticev1.System, spec *latticev1.SystemSpec) (*latticev1.System, error) {
	if reflect.DeepEqual(system.Spec, spec) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Spec = *spec

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
