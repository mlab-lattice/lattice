package servicemesh

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	systembootstrapper.Interface

	WorkloadAnnotations(map[int32]definitionv1.ContainerPort) (map[string]string, error)

	WorkloadAddressAnnotations(*latticev1.Address) (map[string]string, error)

	// TransformWorkloadPodTemplateSpec takes in the PodTemplateSpec generated for a workload, and applies any
	// service mesh related transforms necessary to a copy of the PodTemplateSpec, and returns it.
	TransformWorkloadPodTemplateSpec(
		spec *corev1.PodTemplateSpec,
		namespace string,
		componentPath tree.Path,
		annotations map[string]string,
		ports map[int32]definitionv1.ContainerPort,
	) (*corev1.PodTemplateSpec, error)

	// ServiceMeshPort returns the port the service mesh is listening on for a given component port.
	ServiceMeshPort(annotations map[string]string, port int32) (int32, error)

	// ServiceMeshPorts returns a map whose keys are component ports and values are the port on which the
	// service mesh is listening on for the given key.
	ServiceMeshPorts(annotations map[string]string) (map[int32]int32, error)

	// WorkloadPort returns the component port for a given port that the service mesh is listening on.
	WorkloadPort(annotations map[string]string, port int32) (int32, error)

	// WorkloadPorts returns a map whose keys are service mesh ports and values are the component port for
	// which the service mesh is listening on for the given key.
	WorkloadPorts(annotations map[string]string) (map[int32]int32, error)

	// HasWorkloadIP returns the assigned IP if there is one and the empty string otherwise
	HasWorkloadIP(*latticev1.Address) (string, error)

	// WorkloadIP returns the IP address that should be registered in DNS (assigning one if need be)
	// for the service and annotations that should be applied to the Address.
	WorkloadIP(address *latticev1.Address, workloadPorts map[int32]definitionv1.ContainerPort) (string, map[string]string, error)

	// ReleaseWorkloadIP removes a service IP from the pool of currently leased IPs.
	ReleaseWorkloadIP(*latticev1.Address) (map[string]string, error)

	// IsDeploymentSpecUpdated checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformWorkloadPodTemplateSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformWorkloadPodTemplateSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformWorkloadPodTemplateSpec.
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
