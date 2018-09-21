package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	baseboostrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper/base"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Bootstrap() *cli.Command {
	var (
		latticeID         string
		namespacePrefix   string
		internalDNSDomain string
		kubeConfigPath    string

		containerBuildRegistryAuthType string

		cloudProvider string
	)

	options := &bootstrap.Options{
		Config: latticev1.ConfigSpec{
			ContainerBuild: latticev1.ConfigContainerBuild{
				Builder:        latticev1.ConfigComponentBuildBuilder{},
				DockerArtifact: latticev1.ConfigComponentBuildDockerArtifact{},
			},
		},
		MasterComponents: baseboostrapper.MasterComponentOptions{
			LatticeControllerManager: baseboostrapper.LatticeControllerManagerOptions{},
			APIServer:                baseboostrapper.APIServerOptions{},
		},
	}

	cloudBootstrapFlag, cloudBootstrapOptions := cloudprovider.LatticeBoostrapperFlag(&cloudProvider)

	var serviceMesh string
	serviceMeshBootstrapFlag, serviceMeshBootstrapOptions := servicemesh.LatticeBoostrapperFlag(&serviceMesh)

	var terraformBackend string
	terraformBackendFlag, terraformBackendOptions := terraform.BackendFlags(&terraformBackend)

	var dryRun bool
	var print bool

	return &cli.Command{
		Flags: cli.Flags{
			"lattice-id": &flags.String{
				Required: true,
				Target:   &latticeID,
				Usage:    "ID of the Lattice to bootstrap",
			},

			"namespace-prefix": &flags.String{
				Default: "lattice",
				Target:  &namespacePrefix,
				Usage:   "ID of the Lattice to bootstrap",
			},

			"internal-dns-domain": &flags.String{
				Required: true,
				Target:   &internalDNSDomain,
				Usage:    "dns domain to use for internal domains",
			},

			"kubeconfig": &flags.String{
				Target: &kubeConfigPath,
				Usage:  "path to kubeconfig",
			},

			"api-var": &flags.Embedded{
				Required: true,
				Usage:    "configuration for the api",
				Flags: cli.Flags{
					"image": &flags.String{
						Required: true,
						Target:   &options.MasterComponents.APIServer.Image,
						Usage:    "docker image to user for the api",
					},
					"port": &flags.Int32{
						Default: 80,
						Target:  &options.MasterComponents.APIServer.Port,
						Usage:   "port the api should listen on",
					},
					"host-network": &flags.Bool{
						Target:  &options.MasterComponents.APIServer.HostNetwork,
						Default: false,
						Usage:   "whether or not to run the api on the host network",
					},
					"args": &flags.StringSlice{
						Target: &options.MasterComponents.APIServer.Args,
						Usage:  "extra arguments to pass to the api",
					},
				},
			},

			"controller-manager-var": &flags.Embedded{
				Required: true,
				Usage:    "configuration for the controller manager",
				Flags: cli.Flags{
					"image": &flags.String{
						Required: true,
						Target:   &options.MasterComponents.LatticeControllerManager.Image,
						Usage:    "docker image to user for the controller-manager",
					},
					"args": &flags.StringSlice{
						Target: &options.MasterComponents.LatticeControllerManager.Args,
						Usage:  "extra arguments to pass to the controller manager",
					},
				},
			},

			"container-builder-var": &flags.Embedded{
				Required: true,
				Usage:    "configuration for the container builder",
				Flags: cli.Flags{
					"image": &flags.String{
						Required: true,
						Target:   &options.Config.ContainerBuild.Builder.Image,
						Usage:    "docker image to user for the container-builder",
					},
					"docker-api-version": &flags.String{
						Target: &options.Config.ContainerBuild.Builder.DockerAPIVersion,
						Usage:  "version of the docker API used by the build node docker daemon",
					},
				},
			},

			"container-build-docker-artifact-var": &flags.Embedded{
				Required: true,
				Usage:    "configuration for the docker artifacts produced by the container builder",
				Flags: cli.Flags{
					"registry": &flags.String{
						Target:   &options.Config.ContainerBuild.DockerArtifact.Registry,
						Required: true,
						Usage:    "registry to tag container build docker artifacts with",
					},
					"registry-auth-type": &flags.String{
						Target: &containerBuildRegistryAuthType,
						Usage:  "type of auth to use for the container build registry",
					},
					"repository-per-image": &flags.Bool{
						Target:  &options.Config.ContainerBuild.DockerArtifact.RepositoryPerImage,
						Default: false,
						Usage:   "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used",
					},
					"repository": &flags.String{
						Target: &options.Config.ContainerBuild.DockerArtifact.Repository,
						Usage:  "repository to tag container build docker artifacts with, required if container-build-docker-artifact-repository-per-image is false",
					},
					"push": &flags.Bool{
						Target:  &options.Config.ContainerBuild.DockerArtifact.Push,
						Default: true,
						Usage:   "whether or not the container-builder should push the docker artifact (use false for local)",
					},
				},
			},

			"cloud-provider": &flags.String{
				Required: true,
				Target:   &cloudProvider,
				Usage:    "cloud provider that the kubernetes cluster is running on",
			},
			"cloud-provider-var": cloudBootstrapFlag,

			"service-mesh": &flags.String{
				Required: true,
				Target:   &serviceMesh,
				Usage:    "service mesh to bootstrap the lattice with",
			},
			"service-mesh-var": serviceMeshBootstrapFlag,

			"terraform-backend": &flags.String{
				Required: false,
				Target:   &terraformBackend,
				Usage:    "backend for terraform to use ",
			},
			"terraform-backend-var": terraformBackendFlag,

			"dry-run": &flags.Bool{
				Default: false,
				Target:  &dryRun,
				Usage:   "if set, will not actually bootstrap the cluster. useful with --print",
			},
			"print": &flags.Bool{
				Default: false,
				Target:  &print,
				Usage:   "whether or not to print the resources created or that will be created",
			},
		},
		Run: func(args []string, flags cli.Flags) error {
			latticeID := v1.LatticeID(latticeID)

			var kubeConfig *rest.Config
			if !dryRun {
				var err error
				kubeConfig, err = kubeutil.NewConfig(kubeConfigPath, "")
				if err != nil {
					return err
				}
			}

			options.Terraform = baseboostrapper.TerraformOptions{
				Backend: *terraformBackendOptions,
			}

			if containerBuildRegistryAuthType != "" {
				options.Config.ContainerBuild.DockerArtifact.RegistryAuthType = &containerBuildRegistryAuthType
			}

			cloudBootstrapper, err := cloudprovider.NewLatticeBootstrapper(latticeID, namespacePrefix, internalDNSDomain, cloudBootstrapOptions)
			if err != nil {
				return err
			}

			serviceMeshBootstrapper, err := servicemesh.NewLatticeBootstrapper(namespacePrefix, serviceMeshBootstrapOptions)
			if err != nil {
				return err
			}

			bootstrappers := []bootstrapper.Interface{
				serviceMeshBootstrapper,
				// cloud bootstrapper has to come last so the local bootstrapper can strip taints off of
				// pod specs
				cloudBootstrapper,
			}

			return BootstrapKubernetesLattice(
				v1.LatticeID(latticeID),
				namespacePrefix,
				internalDNSDomain,
				kubeConfig,
				cloudProvider,
				bootstrappers,
				options,
				dryRun,
				print,
			)
		},
	}
}

func BootstrapKubernetesLattice(
	latticeID v1.LatticeID,
	namespacePrefix string,
	internalDNSDomain string,
	kubeConfig *rest.Config,
	cloudProvider string,
	bootstrappers []bootstrapper.Interface,
	options *bootstrap.Options,
	dryRun bool,
	print bool,
) error {
	var kubeClient kubeclientset.Interface
	var latticeClient latticeclientset.Interface

	var resources *bootstrapper.Resources
	var err error
	if dryRun {
		resources, err = bootstrap.GetBootstrapResources(
			latticeID,
			namespacePrefix,
			internalDNSDomain,
			cloudProvider,
			options,
			bootstrappers,
		)
	} else {
		kubeClient, err = kubeclientset.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		latticeClient, err = latticeclientset.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		resources, err = bootstrap.Bootstrap(
			latticeID,
			namespacePrefix,
			internalDNSDomain,
			cloudProvider,
			options,
			bootstrappers,
			kubeConfig,
			kubeClient,
			latticeClient,
		)
	}

	if err != nil {
		return err
	}

	if print {
		resourcesString, err := resources.String()
		if err != nil {
			return err
		}

		fmt.Println(resourcesString)
	}

	return nil
}
