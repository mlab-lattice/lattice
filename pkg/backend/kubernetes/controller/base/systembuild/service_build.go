package systembuild

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/satori/go.uuid"
)

func (c *Controller) createNewServiceBuild(build *crv1.SystemBuild, servicePath tree.NodePath, serviceDefinition definition.Service) (*crv1.ServiceBuild, error) {
	serviceBuild := serviceBuild(build, servicePath, serviceDefinition)
	return c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Create(serviceBuild)
}

func serviceBuild(build *crv1.SystemBuild, servicePath tree.NodePath, serviceDefinition definition.Service) *crv1.ServiceBuild {
	labels := map[string]string{
		constants.LabelKeySystemBuildID:     build.Name,
		constants.LabelKeyServicePathDomain: servicePath.ToDomain(true),
	}

	spec := serviceBuildSpec(serviceDefinition)

	return &crv1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(build, controllerKind)},
		},
		Spec: spec,
		Status: crv1.ServiceBuildStatus{
			State: crv1.ServiceBuildStatePending,
		},
	}
}

func serviceBuildSpec(serviceDefinition definition.Service) crv1.ServiceBuildSpec {
	components := map[string]crv1.ServiceBuildSpecComponentBuildInfo{}
	for _, component := range serviceDefinition.Components() {
		components[component.Name] = crv1.ServiceBuildSpecComponentBuildInfo{
			DefinitionBlock: component.Build,
		}
	}

	return crv1.ServiceBuildSpec{
		Components: components,
	}
}
