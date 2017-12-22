package bootstrap

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	clusterbootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	baseclusterboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper/base"
	systembootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/cli"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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

	cloudProviderName string
	cloudProviderVars []string

	serviceMeshProvider     string
	serviceMeshProviderVars []string

	terraformBackend     string
	terraformBackendVars []string

	networkingProvider     string
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

		cloudProviderConfig, err := parseCloudProviderVars()
		if err != nil {
			panic(err)
		}
		options.Config.CloudProvider = *cloudProviderConfig

		serviceMeshConfig, err := parseServiceMeshVars()
		if err != nil {
			panic(err)
		}
		options.Config.ServiceMesh = *serviceMeshConfig

		terraformConfig, err := parseTerraformVars()
		if err != nil {
			panic(err)
		}
		options.Config.Terraform = terraformConfig

		serviceMesh, err := servicemesh.NewServiceMesh(&options.Config.ServiceMesh)
		if err != nil {
			panic(err)
		}

		cloudProvider, err := cloudprovider.NewCloudProvider(cloudProviderName)
		if err != nil {
			panic(err)
		}

		//networkingOptions, err := parseNetworkingVars()
		//if err != nil {
		//	panic(err)
		//}
		//options.Networking = networkingOptions

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
				cloudProvider,
			)
		} else {
			clusterResources, err = clusterbootstrap.Bootstrap(
				clusterID,
				cloudProviderName,
				options,
				serviceMesh,
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
	Cmd.Flags().BoolVar(&options.Config.ComponentBuild.DockerArtifact.RepositoryPerImage, "component-build-docker-artifact-repository-per-image", false, "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used")
	Cmd.Flags().StringVar(&options.Config.ComponentBuild.DockerArtifact.Repository, "component-build-docker-artifact-repository", "", "repository to tag component build docker artifacts with, required if component-build-docker-artifact-repository-per-image is false")
	Cmd.Flags().BoolVar(&options.Config.ComponentBuild.DockerArtifact.Push, "component-build-docker-artifact-push", true, "whether or not the component-builder should push the docker artifact (use false for local)")

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

	Cmd.Flags().StringVar(&networkingProvider, "networking-provider", "", "provider to use for networking")
	Cmd.Flags().StringArrayVar(&networkingProviderVars, "networking-provider-var", nil, "additional variables for the networking provider")
}

func parseCloudProviderVars() (*crv1.ConfigCloudProvider, error) {
	var config *crv1.ConfigCloudProvider
	switch cloudProviderName {
	case constants.ProviderLocal:
		localConfig, err := parseCloudProviderVarsLocal()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigCloudProvider{
			Local: localConfig,
		}
	case constants.ProviderAWS:
		awsConfig, err := parseProviderCloudVarsAWS()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigCloudProvider{
			AWS: awsConfig,
		}
	default:
		return nil, fmt.Errorf("unsupported cloudProviderName: %v", cloudProviderName)
	}

	return config, nil
}

func parseCloudProviderVarsLocal() (*crv1.ConfigCloudProviderLocal, error) {
	localConfig := &crv1.ConfigCloudProviderLocal{}
	flags := cli.EmbeddedFlag{
		Target: &localConfig,
		Expected: map[string]cli.EmbeddedFlagValue{
			"system-ip": {
				Required:     true,
				EncodingName: "ip",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return localConfig, nil
}

func parseProviderCloudVarsAWS() (*crv1.ConfigCloudProviderAWS, error) {
	awsConfig := &crv1.ConfigCloudProviderAWS{}
	flags := cli.EmbeddedFlag{
		Target: &awsConfig,
		Expected: map[string]cli.EmbeddedFlagValue{
			"region": {
				Required: true,
			},
			"account-id": {
				Required:     true,
				EncodingName: "accountId",
			},
			"vpc-id": {
				Required:     true,
				EncodingName: "vpcId",
			},
			"subnet-ids": {
				Required:     true,
				EncodingName: "subnetIds",
				ValueParser: func(value string) (interface{}, error) {
					return strings.Split(value, ","), nil
				},
			},
			"master-node-security-group-id": {
				Required:     true,
				EncodingName: "masterNodeSecurityGroupId",
			},
			"base-node-ami-id": {
				Required:     true,
				EncodingName: "baseNodeAmiId",
			},
			"key-name": {
				Required:     true,
				EncodingName: "keyName",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return awsConfig, nil
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

func parseTerraformVars() (*crv1.ConfigTerraform, error) {
	if terraformBackend == "" {
		return nil, nil
	}

	var config *crv1.ConfigTerraform
	switch terraformBackend {
	case kubeconstants.TerraformBackendS3:
		s3Config, err := parseTerraformVarsS3()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigTerraform{
			Backend: &crv1.ConfigTerraformBackend{
				S3: s3Config,
			},
		}
	default:
		return nil, fmt.Errorf("unsupported terraform backend: %v", terraformBackend)
	}

	return config, nil
}

func parseTerraformVarsS3() (*crv1.ConfigTerraformBackendS3, error) {
	s3Config := &crv1.ConfigTerraformBackendS3{}
	flags := cli.EmbeddedFlag{
		Target: &s3Config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"bucket": {
				Required: true,
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return s3Config, nil
}

//func parseNetworkingVars() (*cloudboostrapper.NetworkingOptions, error) {
//	if networkingProvider == "" {
//		return nil, nil
//	}
//
//	var options *cloudboostrapper.NetworkingOptions
//	switch terraformBackend {
//	case kubeconstants.NetworkingProviderFlannel:
//		flannelOptions, err := parseNetworkingVarsFlannel()
//		if err != nil {
//			return nil, err
//		}
//		options = &cloudboostrapper.NetworkingOptions{
//			Flannel: flannelOptions,
//		}
//	default:
//		return nil, fmt.Errorf("unsupported networking provider: %v", networkingProvider)
//	}
//
//	return options, nil
//}
//
//func parseNetworkingVarsFlannel() (*cloudboostrapper.FlannelOptions, error) {
//	flannelOptions := &cloudboostrapper.FlannelOptions{}
//	flags := cli.EmbeddedFlag{
//		Target: &flannelOptions,
//		Expected: map[string]cli.EmbeddedFlagValue{
//			"network-cidr-block": {
//				Required:     true,
//				EncodingName: "NetworkCIDRBlock",
//			},
//		},
//	}
//
//	err := flags.Parse(cloudProviderVars)
//	if err != nil {
//		return nil, err
//	}
//	return flannelOptions, nil
//}
