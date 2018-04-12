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

	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Command struct {
}

// Base implements the latticectl.Command interface.
func (c *Command) Base() (*latticectl.BaseCommand, error) {
	var latticeID string
	var kubeConfigPath string

	options := &bootstrap.Options{
		Config: latticev1.ConfigSpec{
			ComponentBuild: latticev1.ConfigComponentBuild{
				Builder:        latticev1.ConfigComponentBuildBuilder{},
				DockerArtifact: latticev1.ConfigComponentBuildDockerArtifact{},
			},
		},
		MasterComponents: baseboostrapper.MasterComponentOptions{
			LatticeControllerManager: baseboostrapper.LatticeControllerManagerOptions{},
			APIServer:                baseboostrapper.APIServerOptions{},
		},
	}
	var componentBuildRegistryAuthType string

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
			&cli.StringFlag{
				Name:    "lattice-id",
				Default: "lattice",
				Target:  &latticeID,
				Usage:   "ID of the Lattice to bootstrap",
			},

			&cli.StringFlag{
				Name:   "kubeconfig",
				Target: &kubeConfigPath,
				Usage:  "path to kubeconfig",
			},

			&cli.EmbeddedFlag{
				Name:     "api-var",
				Required: true,
				Usage:    "configuration for the api",
				Flags: cli.Flags{
					&cli.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.APIServer.Image,
						Usage:    "docker image to user for the api",
					},
					&cli.Int32Flag{
						Name:    "port",
						Default: 80,
						Target:  &options.MasterComponents.APIServer.Port,
						Usage:   "port the api should listen on",
					},
					&cli.BoolFlag{
						Name:    "host-network",
						Target:  &options.MasterComponents.APIServer.HostNetwork,
						Default: false,
						Usage:   "whether or not to run the api on the host network",
					},
					&cli.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.APIServer.Args,
						Usage:  "extra arguments to pass to the api",
					},
				},
			},

			&cli.EmbeddedFlag{
				Name:     "controller-manager-var",
				Required: true,
				Usage:    "configuration for the controller manager",
				Flags: cli.Flags{
					&cli.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.LatticeControllerManager.Image,
						Usage:    "docker image to user for the controller-manager",
					},
					&cli.StringFlag{
						Name:   "terraform-module-path",
						Target: &options.MasterComponents.LatticeControllerManager.TerraformModulePath,
						Usage:  "path to terraform modules",
					},
					&cli.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.LatticeControllerManager.Args,
						Usage:  "extra arguments to pass to the controller manager",
					},
				},
			},

			&cli.EmbeddedFlag{
				Name:     "component-builder-var",
				Required: true,
				Usage:    "configuration for the component builder",
				Flags: cli.Flags{
					&cli.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.Config.ComponentBuild.Builder.Image,
						Usage:    "docker image to user for the component-builder",
					},
					&cli.StringFlag{
						Name:   "docker-api-version",
						Target: &options.Config.ComponentBuild.Builder.DockerAPIVersion,
						Usage:  "version of the docker API used by the build node docker daemon",
					},
				},
			},

			&cli.EmbeddedFlag{
				Name:     "component-build-docker-artifact-var",
				Required: true,
				Usage:    "configuration for the docker artifacts produced by the component builder",
				Flags: cli.Flags{
					&cli.StringFlag{
						Name:     "registry",
						Target:   &options.Config.ComponentBuild.DockerArtifact.Registry,
						Required: true,
						Usage:    "registry to tag component build docker artifacts with",
					},
					&cli.StringFlag{
						Name:   "registry-auth-type",
						Target: &componentBuildRegistryAuthType,
						Usage:  "type of auth to use for the component build registry",
					},
					&cli.BoolFlag{
						Name:    "repository-per-image",
						Target:  &options.Config.ComponentBuild.DockerArtifact.RepositoryPerImage,
						Default: false,
						Usage:   "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used",
					},
					&cli.StringFlag{
						Name:   "repository",
						Target: &options.Config.ComponentBuild.DockerArtifact.Repository,
						Usage:  "repository to tag component build docker artifacts with, required if component-build-docker-artifact-repository-per-image is false",
					},
					&cli.BoolFlag{
						Name:    "push",
						Target:  &options.Config.ComponentBuild.DockerArtifact.Push,
						Default: true,
						Usage:   "whether or not the component-builder should push the docker artifact (use false for local)",
					},
				},
			},

			&cli.StringFlag{
				Name:     "cloud-provider",
				Required: true,
				Target:   &cloudProvider,
				Usage:    "cloud provider that the kubernetes cluster is running on",
			},
			cloudBootstrapFlag,

			&cli.StringFlag{
				Name:     "service-mesh",
				Required: true,
				Target:   &serviceMesh,
				Usage:    "service mesh to bootstrap the lattice with",
			},
			serviceMeshBootstrapFlag,

			&cli.StringFlag{
				Name:     "terraform-backend",
				Required: false,
				Target:   &terraformBackend,
				Usage:    "backend for terraform to use ",
			},
			terraformBackendFlag,

			&cli.BoolFlag{
				Name:    "dry-run",
				Default: false,
				Target:  &dryRun,
				Usage:   "if set, will not actually bootstrap the cluster. useful with --print",
			},
			&cli.BoolFlag{
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

			cloudBootstrapper, err := cloudprovider.NewLatticeBootstrapper(latticeID, cloudBootstrapOptions)
			if err != nil {
				fmt.Printf("error getting cloud bootstrapper: %v", err)
			}

			serviceMeshBootstrapper, err := servicemesh.NewLatticeBootstrapper(latticeID, serviceMeshBootstrapOptions)
			if err != nil {
				fmt.Printf("error getting service mesh bootstrapper: %v", err)
			}

			bootstrappers := []bootstrapper.Interface{
				serviceMeshBootstrapper,
				// cloud bootstrapper has to come last so the local bootstrapper can strip taints off of
				// pod specs
				cloudBootstrapper,
			}

			err = BootstrapKubernetesLattice(v1.LatticeID(latticeID), kubeConfig, cloudProvider, bootstrappers, options, dryRun, print)
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
