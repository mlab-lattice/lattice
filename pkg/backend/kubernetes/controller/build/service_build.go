package build

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/satori/go.uuid"
)

func (c *Controller) createNewServiceBuild(
	build *latticev1.Build,
	servicePath tree.NodePath,
	serviceDefinition definition.Service,
) (*latticev1.ServiceBuild, error) {
	serviceBuild := serviceBuild(build, servicePath, serviceDefinition)
	return c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Create(serviceBuild)
}

func serviceBuild(
	build *latticev1.Build,
	servicePath tree.NodePath,
	serviceDefinition definition.Service,
) *latticev1.ServiceBuild {
	labels := map[string]string{
		latticev1.ServiceBuildPathLabelKey: servicePath.ToDomain(),
	}

	if label, ok := build.DefinitionVersionLabel(); ok {
		labels[latticev1.ServiceBuildDefinitionVersionLabelKey] = label
	}

	spec := serviceBuildSpec(serviceDefinition)

	return &latticev1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*newOwnerReference(build)},
		},
		Spec: spec,
	}
}

func serviceBuildSpec(serviceDefinition definition.Service) latticev1.ServiceBuildSpec {
	components := map[string]latticev1.ServiceBuildSpecComponentBuildInfo{}
	for _, component := range serviceDefinition.Components() {
		components[component.Name] = latticev1.ServiceBuildSpecComponentBuildInfo{
			DefinitionBlock: component.Build,
		}
	}

	return latticev1.ServiceBuildSpec{
		Components: components,
	}
}

func newOwnerReference(build *latticev1.Build) *metav1.OwnerReference {
	gvk := latticev1.BuildKind
	blockOwnerDeletion := true

	// set isController to false, since there should only be one controller
	// owning the lifecycle of the service build. since other builds may also
	// end up adopting the service build, we shouldn't think of any given
	// build as the controller build
	isController := false

	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               build.Name,
		UID:                build.UID,
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
