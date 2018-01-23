package systembuild

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/satori/go.uuid"
)

func (c *Controller) createNewServiceBuild(
	build *latticev1.SystemBuild,
	servicePath tree.NodePath,
	serviceDefinition definition.Service,
) (*latticev1.ServiceBuild, error) {
	serviceBuild := serviceBuild(build, servicePath, serviceDefinition)
	return c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Create(serviceBuild)
}

func serviceBuild(
	build *latticev1.SystemBuild,
	servicePath tree.NodePath,
	serviceDefinition definition.Service,
) *latticev1.ServiceBuild {
	labels := map[string]string{
		constants.LabelKeySystemBuildID:     build.Name,
		constants.LabelKeyServicePathDomain: servicePath.ToDomain(true),
	}

	spec := serviceBuildSpec(serviceDefinition)

	return &latticev1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(build, controllerKind)},
		},
		Spec: spec,
		Status: latticev1.ServiceBuildStatus{
			State: latticev1.ServiceBuildStatePending,
		},
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
