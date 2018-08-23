package servicemesh

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	systembootstrapper.Interface

	ServiceAnnotations(*latticev1.Service) (map[string]string, error)

	ServiceAddressAnnotations(*latticev1.Address) (map[string]string, error)

	// TransformServicePodTemplateSpec takes in the DeploymentSpec generated for a Service, and applies an service mesh
	// related transforms necessary to a copy of the DeploymentSpec, and returns it.
	TransformServicePodTemplateSpec(*latticev1.Service, *corev1.PodTemplateSpec) (*corev1.PodTemplateSpec, error)

	// ServiceMeshPort returns the port the service mesh is listening on for a given component port.
	ServiceMeshPort(*latticev1.Service, int32) (int32, error)

	// ServiceMeshPorts returns a map whose keys are component ports and values are the port on which the
	// service mesh is listening on for the given key.
	ServiceMeshPorts(*latticev1.Service) (map[int32]int32, error)

	// ServicePort returns the component port for a given port that the service mesh is listening on.
	ServicePort(*latticev1.Service, int32) (int32, error)

	// ServicehPorts returns a map whose keys are service mesh ports and values are the component port for
	// which the service mesh is listening on for the given key.
	ServicePorts(*latticev1.Service) (map[int32]int32, error)

	// ServiceHasIP returns the assigned IP if there is one and the empty string otherwise
	HasServiceIP(*latticev1.Address) (string, error)

	// ServiceIP returns the IP address that should be registered in DNS (assigning one if need be)
	// for the service and annotations that should be applied to the Address.
	ServiceIP(*latticev1.Service, *latticev1.Address) (string, map[string]string, error)

	// ReleaseServiceIP removes a service IP from the pool of currently leased IPs.
	ReleaseServiceIP(*latticev1.Address) (map[string]string, error)

	// IsDeploymentSpecUpdated checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformServicePodTemplateSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformServicePodTemplateSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformServicePodTemplateSpec.
	IsDeploymentSpecUpdated(
		service *latticev1.Service,
		current, desired, untransformed *appsv1.DeploymentSpec,
	) (bool, string, *appsv1.DeploymentSpec)
}

type Options struct {
	Envoy *envoy.Options
}

func NewServiceMesh(options *Options) (Interface, error) {
	var serviceMesh Interface
	var err error

	switch {
	case options.Envoy != nil:
		serviceMesh, err = envoy.NewEnvoyServiceMesh(options.Envoy)
	default:
		err = fmt.Errorf("must provide service mesh options")
	}

	return serviceMesh, err
}

func OverlayConfigOptions(staticOptions *Options, dynamicConfig *latticev1.ConfigServiceMesh) (*Options, error) {
	if staticOptions.Envoy != nil {
		if dynamicConfig.Envoy == nil {
			return nil, fmt.Errorf("static options were for envoy but dynamic config did not have envoy options set")
		}

		envoyOptions, err := envoy.NewOptions(staticOptions.Envoy, dynamicConfig.Envoy)
		if err != nil {
			return nil, err
		}

		options := &Options{
			Envoy: envoyOptions,
		}
		return options, nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}

func Flag(serviceMesh *string) (cli.Flag, *Options) {
	envoyFlags, envoyOptions := envoy.Flags()
	options := &Options{}

	flag := &cli.DelayedEmbeddedFlag{
		Name:     "service-mesh-var",
		Required: true,
		Usage:    "configuration for the service mesh",
		Flags: map[string]cli.Flags{
			Envoy: envoyFlags,
		},
		FlagChooser: func() (*string, error) {
			if serviceMesh == nil {
				return nil, fmt.Errorf("service mesh cannot be nil")
			}

			switch *serviceMesh {
			case Envoy:
				options.Envoy = envoyOptions
			default:
				return nil, fmt.Errorf("unsupported service mesh %v", *serviceMesh)
			}

			return serviceMesh, nil
		},
	}

	return flag, options
}
