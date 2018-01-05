package bootstrap

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	awscloudprovider "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	localcloudprovider "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	clusterbootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	baseclusterboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper/base"
	systembootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/cli"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/mlab-lattice/system/pkg/terraform"
	"github.com/spf13/cobra"
)

var (
	printBool bool

	kubeconfigPath    string
	kubeconfigContext string

	defaultLatticeControllerManagerArgs = []string{
		"-v", "5",
		"--logtostderr",
	}

	defaultManagerAPIArgs = []string{}

	clusterIDString string

	initialSystemIDString      string
	initialSystemDefinitionURL string

	componentBuildRegistryAuthType string

	cloudProviderName string
	cloudProviderVars []string

	serviceMeshProvider     string
	serviceMeshProviderVars []string

	terraformBackend     string
	terraformBackendVars []string

	networkingProviderName string
	networkingProviderVars []string
)

var options = &clusterbootstrap.Options{
	Config: crv1.ConfigSpec{
		ComponentBuild: crv1.ConfigComponentBuild{
			Builder:        crv1.ConfigComponentBuildBuilder{},
			DockerArtifact: crv1.ConfigComponentBuildDockerArtifact{},
		},
		ServiceMesh: crv1.ConfigServiceMesh{},
	},
	MasterComponents: baseclusterboostrapper.MasterComponentOptions{
		LatticeControllerManager: baseclusterboostrapper.LatticeControllerManagerOptions{},
		ManagerAPI:               baseclusterboostrapper.ManagerAPIOptions{},
	},
}

var Cmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "bootstraps a kubernetes cluster to run Lattice",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if !options.Config.ComponentBuild.DockerArtifact.RepositoryPerImage && options.Config.ComponentBuild.DockerArtifact.Repository == "" {
			panic("must specify component-build-docker-artifact-repository if not component-build-docker-artifact-repository-per-image")
		}

		if componentBuildRegistryAuthType != "" {
			options.Config.ComponentBuild.DockerArtifact.RegistryAuthType = &componentBuildRegistryAuthType
		}

		emtpy := ""
		if options.Config.ComponentBuild.DockerArtifact.RegistryAuthType == &emtpy {
			options.Config.ComponentBuild.DockerArtifact.RegistryAuthType = nil
		}

		clusterID := types.ClusterID(clusterIDString)
		initialSystemID := types.SystemID(initialSystemIDString)

		var kubeConfig *rest.Config
		if !options.DryRun {
			var err error
			kubeConfig, err = kubeutil.NewConfig(kubeconfigPath, "")
			if err != nil {
				panic(err)
			}
		}

		cloudProviderOptions, err := parseCloudProviderVars()
		if err != nil {
			panic(err)
		}

		serviceMeshConfig, err := parseServiceMeshVars()
		if err != nil {
			panic(err)
		}
		options.Config.ServiceMesh = *serviceMeshConfig

		terraformOptions, err := parseTerraformVars()
		if err != nil {
			panic(err)
		}
		options.Terraform = terraformOptions

		serviceMesh, err := servicemesh.NewServiceMesh(&options.Config.ServiceMesh)
		if err != nil {
			panic(err)
		}

		cloudProvider, err := cloudprovider.NewCloudProvider(cloudProviderOptions)
		if err != nil {
			panic(err)
		}

		var networkingProvider networkingprovider.Interface
		if networkingProviderName != "" {
			networkingProviderOptions, err := parseNetworkingVars()
			if err != nil {
				panic(err)
			}

			networkingProvider, err = networkingprovider.NewNetworkingProvider(networkingProviderOptions)
			if err != nil {
				panic(err)
			}
		}

		var kubeClient kubeclientset.Interface
		var latticeClient latticeclientset.Interface

		if !options.DryRun {
			kubeClient = kubeclientset.NewForConfigOrDie(kubeConfig)
			latticeClient = latticeclientset.NewForConfigOrDie(kubeConfig)
		}

		var clusterResources *clusterbootstrapper.ClusterResources
		if options.DryRun {
			clusterResources, err = clusterbootstrap.GetBootstrapResources(
				clusterID,
				cloudProviderName,
				options,
				serviceMesh,
				networkingProvider,
				cloudProvider,
			)
		} else {
			clusterResources, err = clusterbootstrap.Bootstrap(
				clusterID,
				cloudProviderName,
				options,
				serviceMesh,
				networkingProvider,
				cloudProvider,
				kubeConfig,
				kubeClient,
				latticeClient,
			)
		}

		if err != nil {
			panic(err)
		}

		if printBool {
			resourcesString, err := clusterResources.String()
			if err != nil {
				panic(err)
			}

			fmt.Println(resourcesString)
		}

		if initialSystemDefinitionURL == "" {
			return
		}

		var systemResources *systembootstrapper.SystemResources
		if options.DryRun {
			systemResources = systembootstrap.GetBootstrapResources(clusterID, initialSystemID, initialSystemDefinitionURL, serviceMesh, cloudProvider)
		} else {
			fmt.Printf("bootstrapping initial system \"%v\"\n", initialSystemIDString)
			systemResources, err = systembootstrap.Bootstrap(
				clusterID,
				initialSystemID,
				initialSystemDefinitionURL,
				serviceMesh,
				cloudProvider,
				kubeClient,
				latticeClient,
			)
		}

		if err != nil {
			panic(err)
		}

		if printBool {
			resourcesString, err := systemResources.String()
			if err != nil {
				panic(err)
			}

			fmt.Println(resourcesString)
		}
	},
}

func init() {
	Cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "if set, will not actually bootstrap the cluster. useful with --printBool")
	Cmd.Flags().BoolVar(&printBool, "print", false, "whether or not to printBool the resources created or that will be created")
	Cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig")
	Cmd.Flags().StringVar(&kubeconfigContext, "kubeconfig-context", "", "context in the kubeconfig to use")

	Cmd.Flags().StringVar(&clusterIDString, "cluster-id", "lattice", "lattice cluster ID")

	Cmd.Flags().StringVar(&options.Config.ComponentBuild.Builder.Image, "component-builder-image", "", "docker image to user for the component-builder")
	Cmd.MarkFlagRequired("component-builder-image")
	Cmd.Flags().StringVar(&options.Config.ComponentBuild.Builder.DockerAPIVersion, "component-builder-docker-api-version", "", "version of the docker API used by the build node docker daemon")

	Cmd.Flags().StringVar(&options.Config.ComponentBuild.DockerArtifact.Registry, "component-build-docker-artifact-registry", "", "registry to tag component build docker artifacts with")
	Cmd.MarkFlagRequired("component-build-docker-artifact-registry")
	Cmd.Flags().StringVar(&componentBuildRegistryAuthType, "component-build-docker-artifact-registry-auth-type", "", "type of auth to use for the component build registry")
	Cmd.Flags().BoolVar(&options.Config.ComponentBuild.DockerArtifact.RepositoryPerImage, "component-build-docker-artifact-repository-per-image", false, "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used")
	Cmd.Flags().StringVar(&options.Config.ComponentBuild.DockerArtifact.Repository, "component-build-docker-artifact-repository", "", "repository to tag component build docker artifacts with, required if component-build-docker-artifact-repository-per-image is false")
	Cmd.Flags().BoolVar(&options.Config.ComponentBuild.DockerArtifact.Push, "component-build-docker-artifact-push", true, "whether or not the component-builder should push the docker artifact (use false for localcloudprovider)")

	Cmd.Flags().StringVar(&options.MasterComponents.LatticeControllerManager.Image, "lattice-controller-manager-image", "", "docker image to user for the lattice-controller-manager")
	Cmd.MarkFlagRequired("lattice-controller-manager-image")
	Cmd.Flags().StringArrayVar(&options.MasterComponents.LatticeControllerManager.Args, "lattice-controller-manager-args", defaultLatticeControllerManagerArgs, "extra arguments (besides --cloudProviderName) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&options.MasterComponents.ManagerAPI.Image, "manager-api-image", "", "docker image to user for the lattice-controller-manager")
	Cmd.MarkFlagRequired("manager-api-image")
	Cmd.Flags().Int32Var(&options.MasterComponents.ManagerAPI.Port, "manager-api-port", 80, "port that the manager-api should listen on")
	Cmd.Flags().BoolVar(&options.MasterComponents.ManagerAPI.HostNetwork, "manager-api-host-network", true, "whether or not the manager-api should be on the host network")
	Cmd.Flags().StringArrayVar(&options.MasterComponents.ManagerAPI.Args, "manager-api-args", defaultManagerAPIArgs, "extra arguments (besides --cloudProviderName) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&initialSystemIDString, "initial-system-name", "default", "name to use for the initial system if --initial-system-definition-url is set")
	Cmd.Flags().StringVar(&initialSystemDefinitionURL, "initial-system-definition-url", "", "URL to use for the definition of the optional initial system")

	Cmd.Flags().StringVar(&cloudProviderName, "cloud-provider", "", "cloud provider that the cluster is being bootstrapped on")
	Cmd.MarkFlagRequired("cloud-provider")
	Cmd.Flags().StringArrayVar(&cloudProviderVars, "cloud-provider-var", nil, "additional variables for the cloud provider")

	Cmd.Flags().StringVar(&serviceMeshProvider, "service-mesh", "", "service mesh provider to use")
	Cmd.MarkFlagRequired("service-provider")
	Cmd.Flags().StringArrayVar(&serviceMeshProviderVars, "service-mesh-var", nil, "additional variables for the cloud provider")

	Cmd.Flags().StringVar(&terraformBackend, "terraform-backend", "", "backend to use for terraform")
	Cmd.Flags().StringArrayVar(&terraformBackendVars, "terraform-backend-var", nil, "additional variables for the terraform backend")

	Cmd.Flags().StringVar(&networkingProviderName, "networking-provider", "", "provider to use for networking")
	Cmd.Flags().StringArrayVar(&networkingProviderVars, "networking-provider-var", nil, "additional variables for the networking provider")
}

func parseCloudProviderVars() (*cloudprovider.Options, error) {
	var options *cloudprovider.Options
	switch cloudProviderName {
	case constants.ProviderLocal:
		localOptions, err := parseCloudProviderVarsLocal()
		if err != nil {
			return nil, err
		}
		options = &cloudprovider.Options{
			Local: localOptions,
		}

	case constants.ProviderAWS:
		awsOptions, err := parseProviderCloudVarsAWS()
		if err != nil {
			return nil, err
		}
		options = &cloudprovider.Options{
			AWS: awsOptions,
		}

	default:
		return nil, fmt.Errorf("unsupported cloudProviderName: %v", cloudProviderName)
	}

	return options, nil
}

func parseCloudProviderVarsLocal() (*localcloudprovider.Options, error) {
	options := &localcloudprovider.Options{}
	flags := cli.EmbeddedFlag{
		Target: &options,
		Expected: map[string]cli.EmbeddedFlagValue{
			"cluster-ip": {
				Required:     true,
				EncodingName: "IP",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func parseProviderCloudVarsAWS() (*awscloudprovider.Options, error) {
	awsOptions := &awscloudprovider.Options{}
	flags := cli.EmbeddedFlag{
		Target: &awsOptions,
		Expected: map[string]cli.EmbeddedFlagValue{
			"region": {
				Required: true,
			},
			"account-id": {
				Required:     true,
				EncodingName: "AccountID",
			},
			"vpc-id": {
				Required:     true,
				EncodingName: "VPCID",
			},
			"route53-private-zone-id": {
				Required:     true,
				EncodingName: "Route53PrivateZoneID",
			},
			"subnet-ids": {
				Required:     true,
				EncodingName: "SubnetIDs",
				ValueParser: func(value string) (interface{}, error) {
					return strings.Split(value, ","), nil
				},
			},
			"master-node-security-group-id": {
				Required:     true,
				EncodingName: "MasterNodeSecurityGroupID",
			},
			"base-node-ami-id": {
				Required:     true,
				EncodingName: "BaseNodeAMIID",
			},
			"key-name": {
				Required:     true,
				EncodingName: "KeyName",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return awsOptions, nil
}

func parseServiceMeshVars() (*crv1.ConfigServiceMesh, error) {
	var config *crv1.ConfigServiceMesh
	switch serviceMeshProvider {
	case constants.ServiceMeshEnvoy:
		envoyConfig, err := parseServiceMeshVarsEnvoy()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigServiceMesh{
			Envoy: envoyConfig,
		}
	default:
		return nil, fmt.Errorf("unsupported service mesh provider: %v", serviceMeshProvider)
	}

	return config, nil
}

func parseServiceMeshVarsEnvoy() (*crv1.ConfigEnvoy, error) {
	envoyConfig := &crv1.ConfigEnvoy{}
	flags := cli.EmbeddedFlag{
		Target: &envoyConfig,
		Expected: map[string]cli.EmbeddedFlagValue{
			"prepare-image": {
				Required:     true,
				EncodingName: "PrepareImage",
			},
			"envoy-image": {
				Default:      "envoyproxy/envoy-alpine",
				EncodingName: "Image",
			},
			"redirect-cidr-block": {
				Required:     true,
				EncodingName: "RedirectCIDRBlock",
			},
			"xds-api-image": {
				Required:     true,
				EncodingName: "XDSAPIImage",
			},
			"xds-api-port": {
				Default:      8080,
				EncodingName: "XDSAPIPort",
			},
		},
	}

	err := flags.Parse(serviceMeshProviderVars)
	if err != nil {
		return nil, err
	}
	return envoyConfig, nil
}

func parseTerraformVars() (baseclusterboostrapper.TerraformOptions, error) {
	if terraformBackend == "" {
		return baseclusterboostrapper.TerraformOptions{}, nil
	}

	var backend terraform.BackendOptions
	switch terraformBackend {
	case terraform.BackendS3:
		s3Config, err := parseTerraformVarsS3()
		if err != nil {
			return baseclusterboostrapper.TerraformOptions{}, err
		}
		backend = terraform.BackendOptions{
			S3: s3Config,
		}

	default:
		return baseclusterboostrapper.TerraformOptions{}, fmt.Errorf("unsupported terraform backend: %v", terraformBackend)
	}

	options := baseclusterboostrapper.TerraformOptions{
		Backend: backend,
	}
	return options, nil
}

func parseTerraformVarsS3() (*terraform.BackendOptionsS3, error) {
	s3Config := &terraform.BackendOptionsS3{}
	flags := cli.EmbeddedFlag{
		Target: &s3Config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"bucket": {
				EncodingName: "Bucket",
				Required:     true,
			},
		},
	}

	err := flags.Parse(terraformBackendVars)
	if err != nil {
		return nil, err
	}
	return s3Config, nil
}

func parseNetworkingVars() (*networkingprovider.Options, error) {
	var options *networkingprovider.Options
	switch networkingProviderName {
	case networkingprovider.Flannel:
		flannelOptions, err := parseNetworkingVarsFlannel()
		if err != nil {
			return nil, err
		}
		options = &networkingprovider.Options{
			Flannel: flannelOptions,
		}
	default:
		return nil, fmt.Errorf("unsupported networking provider: %v", networkingProviderName)
	}

	return options, nil
}

func parseNetworkingVarsFlannel() (*flannel.Options, error) {
	flannelOptions := &flannel.Options{}
	flags := cli.EmbeddedFlag{
		Target: &flannelOptions,
		Expected: map[string]cli.EmbeddedFlagValue{
			"cidr-block": {
				Required:     true,
				EncodingName: "CIDRBlock",
			},
		},
	}

	err := flags.Parse(networkingProviderVars)
	if err != nil {
		return nil, err
	}
	return flannelOptions, nil
}
