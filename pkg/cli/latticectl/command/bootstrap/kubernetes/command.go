package kubernetes

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	clusterbootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	baseclusterboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper/base"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Command struct {
	Subcommands []latticectl.Command
}

// Base implements the latticectl.Command interface.
func (c *Command) Base() (*latticectl.BaseCommand, error) {
	var latticeID string
	var kubeConfigPath string

	options := &clusterbootstrap.Options{
		Config: latticev1.ConfigSpec{
			ComponentBuild: latticev1.ConfigComponentBuild{
				Builder:        latticev1.ConfigComponentBuildBuilder{},
				DockerArtifact: latticev1.ConfigComponentBuildDockerArtifact{},
			},
		},
		MasterComponents: baseclusterboostrapper.MasterComponentOptions{
			LatticeControllerManager: baseclusterboostrapper.LatticeControllerManagerOptions{},
			ManagerAPI:               baseclusterboostrapper.ManagerAPIOptions{},
		},
	}
	var componentBuildRegistryAuthType string

	var cloudProvider string
	cloudBootstrapFlag, cloudBootstrapOptions := cloudprovider.ClusterBoostrapperFlag(&cloudProvider)

	var serviceMesh string
	serviceMeshBootstrapFlag, serviceMeshBootstrapOptions := servicemesh.ClusterBoostrapperFlag(&serviceMesh)

	var dryRun bool
	var print bool

	cmd := &latticectl.BaseCommand{
		Name: "kubernetes",
		Flags: command.Flags{
			&command.StringFlag{
				Name:    "lattice-id",
				Default: "lattice",
				Target:  &latticeID,
				Usage:   "ID of the Lattice to bootstrap",
			},

			&command.StringFlag{
				Name:   "kubeconfig",
				Target: &kubeConfigPath,
				Usage:  "path to kubeconfig",
			},

			&command.EmbeddedFlag{
				Name:     "component-builder-var",
				Required: true,
				Usage:    "configuration for the component builder",
				Flags: command.Flags{
					&command.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.Config.ComponentBuild.Builder.Image,
						Usage:    "docker image to user for the component-builder",
					},
					&command.StringFlag{
						Name:   "docker-api-version",
						Target: &options.Config.ComponentBuild.Builder.DockerAPIVersion,
						Usage:  "version of the docker API used by the build node docker daemon",
					},
				},
			},

			&command.EmbeddedFlag{
				Name:     "component-build-docker-artifact-var",
				Required: true,
				Usage:    "configuration for the docker artifacts produced by the component builder",
				Flags: command.Flags{
					&command.StringFlag{
						Name:     "registry",
						Target:   &options.Config.ComponentBuild.DockerArtifact.Registry,
						Required: true,
						Usage:    "registry to tag component build docker artifacts with",
					},
					&command.StringFlag{
						Name:   "registry-auth-type",
						Target: &componentBuildRegistryAuthType,
						Usage:  "type of auth to use for the component build registry",
					},
					&command.BoolFlag{
						Name:    "repository-per-image",
						Target:  &options.Config.ComponentBuild.DockerArtifact.RepositoryPerImage,
						Default: false,
						Usage:   "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used",
					},
					&command.StringFlag{
						Name:   "repository",
						Target: &options.Config.ComponentBuild.DockerArtifact.Repository,
						Usage:  "repository to tag component build docker artifacts with, required if component-build-docker-artifact-repository-per-image is false",
					},
					&command.BoolFlag{
						Name:    "push",
						Target:  &options.Config.ComponentBuild.DockerArtifact.Push,
						Default: true,
						Usage:   "whether or not the component-builder should push the docker artifact (use false for local)",
					},
				},
			},

			&command.EmbeddedFlag{
				Name:     "controller-manager-var",
				Required: true,
				Usage:    "configuration for the controller manager",
				Flags: command.Flags{
					&command.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.LatticeControllerManager.Image,
						Usage:    "docker image to user for the controller-manager",
					},
					&command.StringFlag{
						Name:   "terraform-module-path",
						Target: &options.MasterComponents.LatticeControllerManager.TerraformModulePath,
						Usage:  "path to terraform modules",
					},
					&command.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.LatticeControllerManager.Args,
						Usage:  "extra arguments to pass to the controller manager",
					},
				},
			},

			&command.EmbeddedFlag{
				Name:     "api-var",
				Required: true,
				Usage:    "configuration for the api",
				Flags: command.Flags{
					&command.StringFlag{
						Name:     "image",
						Required: true,
						Target:   &options.MasterComponents.ManagerAPI.Image,
						Usage:    "docker image to user for the api",
					},
					&command.Int32Flag{
						Name:    "port",
						Default: 80,
						Target:  &options.MasterComponents.ManagerAPI.Port,
						Usage:   "port the api should listen on",
					},
					&command.BoolFlag{
						Name:   "host-network",
						Target: &options.MasterComponents.ManagerAPI.HostNetwork,
						// TODO: this used to be true
						Default: false,
						Usage:   "whether or not to run the api on the host network",
					},
					&command.StringSliceFlag{
						Name:   "args",
						Target: &options.MasterComponents.ManagerAPI.Args,
						Usage:  "extra arguments to pass to the api",
					},
				},
			},

			&command.StringFlag{
				Name:     "cloud-provider",
				Required: true,
				Target:   &cloudProvider,
				Usage:    "cloud provider that the kubernetes cluster is running on",
			},
			cloudBootstrapFlag,

			&command.StringFlag{
				Name:     "service-mesh",
				Required: true,
				Target:   &serviceMesh,
				Usage:    "service mesh to bootstrap the lattice with",
			},
			serviceMeshBootstrapFlag,

			&command.BoolFlag{
				Name:    "dry-run",
				Default: false,
				Target:  &dryRun,
				Usage:   "if set, will not actually bootstrap the cluster. useful with --print",
			},
			&command.BoolFlag{
				Name:    "print",
				Default: false,
				Target:  &print,
				Usage:   "whether or not to print the resources created or that will be created",
			},
		},
		Run: func(latticectl *latticectl.Latticectl, args []string) {
			latticeID := types.LatticeID(latticeID)

			var kubeConfig *rest.Config
			if !dryRun {
				var err error
				kubeConfig, err = kubeutil.NewConfig(kubeConfigPath, "")
				if err != nil {
					fmt.Printf("error getting kube config: %v", kubeConfig)
				}
			}

			cloudBootstrapper, err := cloudprovider.NewClusterBootstrapper(latticeID, cloudBootstrapOptions)
			if err != nil {
				fmt.Printf("error getting cloud bootstrapper: %v", err)
			}

			serviceMeshBootstrapper, err := servicemesh.NewClusterBootstrapper(serviceMeshBootstrapOptions)
			if err != nil {
				fmt.Printf("error getting service mesh bootstrapper: %v", err)
			}

			bootstrappers := []bootstrapper.Interface{
				serviceMeshBootstrapper,
				// cloud bootstrapper has to come last so the local bootstrapper can strip taints off of
				// pod specs
				cloudBootstrapper,
			}

			err = BootstrapKubernetesLattice(types.LatticeID(latticeID), kubeConfig, cloudProvider, bootstrappers, options, dryRun, print)
			if err != nil {
				fmt.Printf("error bootstrapping lattice: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd, nil
}

func BootstrapKubernetesLattice(
	latticeID types.LatticeID,
	kubeConfig *rest.Config,
	cloudProvider string,
	bootstrappers []bootstrapper.Interface,
	options *clusterbootstrap.Options,
	dryRun bool,
	print bool,
) error {
	var kubeClient kubeclientset.Interface
	var latticeClient latticeclientset.Interface

	var resources *clusterbootstrapper.ClusterResources
	var err error
	if dryRun {
		resources, err = clusterbootstrap.GetBootstrapResources(
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

		resources, err = clusterbootstrap.Bootstrap(
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
