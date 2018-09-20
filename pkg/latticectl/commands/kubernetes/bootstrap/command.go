package bootstrap

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	baseboostrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper/base"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"

	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Command struct {
}

// Base implements the latticectl.Command interface.
func (c *Command) Base() (*latticectl.BaseCommand, error) {
	var latticeID string
	var namespacePrefix string
	var internalDNSDomain string
	var kubeConfigPath string

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
	var containerBuildRegistryAuthType string

	var cloudProvider string
	cloudBootstrapFlag, cloudBootstrapOptions := cloudprovider.LatticeBoostrapperFlag(&cloudProvider)

	var serviceMesh string
	serviceMeshBootstrapFlag, serviceMeshBootstrapOptions := servicemesh.LatticeBoostrapperFlag(&serviceMesh)

	var terraformBackend string
	terraformBackendFlag, terraformBackendOptions := terraform.BackendFlags(&terraformBackend)

	var dryRun bool
	var print bool

	cmd := &latticectl.BaseCommand{
		Name: "bootstrap",
		Flags: cli.Flags{
			&flags.String{
				Name:     "lattice-id",
				Required: true,
				Target:   &latticeID,
				Usage:    "ID of the Lattice to bootstrap",
			},

			&flags.String{
				Name:    "namespace-prefix",
				Default: "lattice",
				Target:  &namespacePrefix,
				Usage:   "ID of the Lattice to bootstrap",
			},

			&flags.String{
				Name:     "internal-dns-domain",
				Required: true,
				Target:   &internalDNSDomain,
				Usage:    "dns domain to use for internal domains",
			},

			&flags.String{
				Name:   "kubeconfig",
				Target: &kubeConfigPath,
				Usage:  "path to kubeconfig",
			},

			&flags.Embedded{
				Name:     "api-var",
				Required: true,
				Usage:    "configuration for the api",
				Flags: cli.Flags{
					&flags.String{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.APIServer.Image,
						Usage:    "docker image to user for the api",
					},
					&flags.Int32{
						Name:    "port",
						Default: 80,
						Target:  &options.MasterComponents.APIServer.Port,
						Usage:   "port the api should listen on",
					},
					&flags.Bool{
						Name:    "host-network",
						Target:  &options.MasterComponents.APIServer.HostNetwork,
						Default: false,
						Usage:   "whether or not to run the api on the host network",
					},
					&flags.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.APIServer.Args,
						Usage:  "extra arguments to pass to the api",
					},
				},
			},

			&flags.Embedded{
				Name:     "controller-manager-var",
				Required: true,
				Usage:    "configuration for the controller manager",
				Flags: cli.Flags{
					&flags.String{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.LatticeControllerManager.Image,
						Usage:    "docker image to user for the controller-manager",
					},
					&flags.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.LatticeControllerManager.Args,
						Usage:  "extra arguments to pass to the controller manager",
					},
				},
			},

			&flags.Embedded{
				Name:     "container-builder-var",
				Required: true,
				Usage:    "configuration for the container builder",
				Flags: cli.Flags{
					&flags.String{
						Name:     "image",
						Required: true,
						Target:   &options.Config.ContainerBuild.Builder.Image,
						Usage:    "docker image to user for the container-builder",
					},
					&flags.String{
						Name:   "docker-api-version",
						Target: &options.Config.ContainerBuild.Builder.DockerAPIVersion,
						Usage:  "version of the docker API used by the build node docker daemon",
					},
				},
			},

			&flags.Embedded{
				Name:     "container-build-docker-artifact-var",
				Required: true,
				Usage:    "configuration for the docker artifacts produced by the container builder",
				Flags: cli.Flags{
					&flags.String{
						Name:     "registry",
						Target:   &options.Config.ContainerBuild.DockerArtifact.Registry,
						Required: true,
						Usage:    "registry to tag container build docker artifacts with",
					},
					&flags.String{
						Name:   "registry-auth-type",
						Target: &containerBuildRegistryAuthType,
						Usage:  "type of auth to use for the container build registry",
					},
					&flags.Bool{
						Name:    "repository-per-image",
						Target:  &options.Config.ContainerBuild.DockerArtifact.RepositoryPerImage,
						Default: false,
						Usage:   "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used",
					},
					&flags.String{
						Name:   "repository",
						Target: &options.Config.ContainerBuild.DockerArtifact.Repository,
						Usage:  "repository to tag container build docker artifacts with, required if container-build-docker-artifact-repository-per-image is false",
					},
					&flags.Bool{
						Name:    "push",
						Target:  &options.Config.ContainerBuild.DockerArtifact.Push,
						Default: true,
						Usage:   "whether or not the container-builder should push the docker artifact (use false for local)",
					},
				},
			},

			&flags.String{
				Name:     "cloud-provider",
				Required: true,
				Target:   &cloudProvider,
				Usage:    "cloud provider that the kubernetes cluster is running on",
			},
			cloudBootstrapFlag,

			&flags.String{
				Name:     "service-mesh",
				Required: true,
				Target:   &serviceMesh,
				Usage:    "service mesh to bootstrap the lattice with",
			},
			serviceMeshBootstrapFlag,

			&flags.String{
				Name:     "terraform-backend",
				Required: false,
				Target:   &terraformBackend,
				Usage:    "backend for terraform to use ",
			},
			terraformBackendFlag,

			&flags.Bool{
				Name:    "dry-run",
				Default: false,
				Target:  &dryRun,
				Usage:   "if set, will not actually bootstrap the cluster. useful with --print",
			},
			&flags.Bool{
				Name:    "print",
				Default: false,
				Target:  &print,
				Usage:   "whether or not to print the resources created or that will be created",
			},
		},
		Run: func(latticectl *latticectl.Latticectl, args []string) {
			latticeID := v1.LatticeID(latticeID)

			var kubeConfig *rest.Config
			if !dryRun {
				var err error
				kubeConfig, err = kubeutil.NewConfig(kubeConfigPath, "")
				if err != nil {
					fmt.Printf("error getting kube config: %v", kubeConfig)
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
				fmt.Printf("error getting cloud bootstrapper: %v", err)
			}

			serviceMeshBootstrapper, err := servicemesh.NewLatticeBootstrapper(namespacePrefix, serviceMeshBootstrapOptions)
			if err != nil {
				fmt.Printf("error getting service mesh bootstrapper: %v", err)
			}

			bootstrappers := []bootstrapper.Interface{
				serviceMeshBootstrapper,
				// cloud bootstrapper has to come last so the local bootstrapper can strip taints off of
				// pod specs
				cloudBootstrapper,
			}

			err = BootstrapKubernetesLattice(
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
			if err != nil {
				fmt.Printf("error bootstrapping lattice: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd, nil
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
