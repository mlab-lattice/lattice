package servicemesh

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
)

type Interface interface {
	TransformServiceDeploymentSpec(*crv1.Service, *appsv1.DeploymentSpec) *appsv1.DeploymentSpec
}

func NewServiceMesh(config *crv1.ConfigServiceMesh) (Interface, error) {
	if config.Envoy != nil {
		return envoy.NewEnvoyServiceMesh(config.Envoy), nil
	}

	return nil, fmt.Errorf("no service mesh configuration set")
}
