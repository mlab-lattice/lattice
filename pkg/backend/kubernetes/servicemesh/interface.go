package servicemesh

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
)

type Interface interface {
	// TransformServiceDeploymentSpec takes in the DeploymentSpec generated for a Service, and applies an service mesh
	// related transforms necessary to a copy of the DeploymentSpec, and returns it.
	TransformServiceDeploymentSpec(*crv1.Service, *appsv1.DeploymentSpec) *appsv1.DeploymentSpec

	// IsDeploymentSpecUpdated checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformServiceDeploymentSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformServiceDeploymentSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformServiceDeploymentSpec.
	IsDeploymentSpecUpdated(service *crv1.Service, current, desired, untransformed *appsv1.DeploymentSpec) (bool, string, *appsv1.DeploymentSpec)
}

func NewServiceMesh(config *crv1.ConfigServiceMesh) (Interface, error) {
	if config.Envoy != nil {
		return envoy.NewEnvoyServiceMesh(config.Envoy), nil
	}

	return nil, fmt.Errorf("no service mesh configuration set")
}
