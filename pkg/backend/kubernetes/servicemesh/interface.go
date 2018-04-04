package servicemesh

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"

	appsv1 "k8s.io/api/apps/v1"
)

type Interface interface {
	ServiceAnnotations(*latticev1.Service) (map[string]string, error)

	// TransformServiceDeploymentSpec takes in the DeploymentSpec generated for a Service, and applies an service mesh
	// related transforms necessary to a copy of the DeploymentSpec, and returns it.
	TransformServiceDeploymentSpec(*latticev1.Service, *appsv1.DeploymentSpec) (*appsv1.DeploymentSpec, error)

	// ServiceMeshPort returns the port the service mesh is listening on for a given component port.
	ServiceMeshPort(*latticev1.Service, int32) (int32, error)

	// ServiceMeshPorts returns a map whose keys are component ports and values are the port on which the
	// service mesh is listening on for the given key.
	ServiceMeshPorts(*latticev1.Service) (map[int32]int32, error)

	// ServicePort returns the component port for a given port that the service mesh is listening on.
	ServicePort(*latticev1.Service, int32) (int32, error)

	// ServiceMeshPorts returns a map whose keys are service mesh ports and values are the component port for
	// which the service mesh is listening on for the given key.
	ServicePorts(*latticev1.Service) (map[int32]int32, error)

	// IsDeploymentSpecUpdated checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformServiceDeploymentSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformServiceDeploymentSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformServiceDeploymentSpec.
	IsDeploymentSpecUpdated(
		service *latticev1.Service,
		current, desired, untransformed *appsv1.DeploymentSpec,
	) (bool, string, *appsv1.DeploymentSpec)

	GetEndpointSpec(*latticev1.ServiceAddress) (*latticev1.EndpointSpec, error)
}

type Options struct {
	Envoy *envoy.Options
}

func OptionsFromConfig(config *latticev1.ConfigServiceMesh) (*Options, error) {
	if config.Envoy != nil {
		options := &Options{
			Envoy: &envoy.Options{
				PrepareImage:      config.Envoy.PrepareImage,
				Image:             config.Envoy.Image,
				RedirectCIDRBlock: config.Envoy.RedirectCIDRBlock,
				XDSAPIPort:        config.Envoy.XDSAPIPort,
			},
		}

		return options, nil
	}

	return nil, fmt.Errorf("must provide service mesh config")
}

func NewServiceMesh(options *Options) (Interface, error) {
	if options.Envoy != nil {
		return envoy.NewEnvoyServiceMesh(options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}
