package bootstrap

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	awscloudprovider "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	localcloudprovider "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	clusterbootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	baseclusterboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper/base"
	systembootstrap "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/none"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/cli"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"
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

// FIXME :: temporary until better solution for nested struct.
type localCloudOptionsFlat struct {
	IP                 string
	DNSControllerImage string
	DnsmasqNannyImage  string
	DnsmasqNannyArgs   []string
	DNSControllerArgs  []string
}

var Cmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "bootstraps a kubernetes cluster to run Client",
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

		cloudProviderClusterOptions, cloudProviderSystemOptions, err := parseCloudProviderVars()
		if err != nil {
			panic(err)
		}

		serviceMeshClusterOptions, serviceMeshSystemOptions, err := parseServiceMeshVars()
		if err != nil {
			panic(err)
		}

		terraformOptions, err := parseTerraformVars()
		if err != nil {
			panic(err)
		}
		options.Terraform = terraformOptions

		cloudProviderClusterBootstrapper, err := cloudprovider.NewClusterBootstrapper(clusterID, cloudProviderClusterOptions)
		if err != nil {
			panic(err)
		}

		serviceMeshClusterBootstrapper, err := servicemesh.NewClusterBootstrapper(serviceMeshClusterOptions)
		if err != nil {
			panic(err)
		}

		clusterBootstrappers := []clusterbootstrapper.Interface{
			serviceMeshClusterBootstrapper,
		}

		var networkingProviderSystemOptions *networkingprovider.SystemBootstrapperOptions
		if networkingProviderName != "" {
			var networkingProviderClusterOptions *networkingprovider.ClusterBootstrapperOptions
			networkingProviderClusterOptions, networkingProviderSystemOptions, err = parseNetworkingVars()
			if err != nil {
				panic(err)
			}

			networkingProviderClusterBootstrapper, err := networkingprovider.NewClusterBootstrapper(networkingProviderClusterOptions)
			if err != nil {
				panic(err)
			}

			clusterBootstrappers = append(clusterBootstrappers, networkingProviderClusterBootstrapper)
		}

		// cloud bootstrapper has to come last
		clusterBootstrappers = append(clusterBootstrappers, cloudProviderClusterBootstrapper)

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
				clusterBootstrappers,
			)
		} else {
			clusterResources, err = clusterbootstrap.Bootstrap(
				clusterID,
				cloudProviderName,
				options,
				clusterBootstrappers,
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

		cloudProviderSystemBootstrapper, err := cloudprovider.NewSystemBootstrapper(cloudProviderSystemOptions)
		if err != nil {
			panic(err)
		}

		serviceMeshSystemBootstrapper, err := servicemesh.NewSystemBootstrapper(serviceMeshSystemOptions)
		if err != nil {
			panic(err)
		}

		systemBootstrappers := []systembootstrapper.Interface{
			serviceMeshSystemBootstrapper,
		}

		if networkingProviderName != "" {
			networkingProviderSystemBootstrapper, err := networkingprovider.NewSystemBootstrapper(networkingProviderSystemOptions)
			if err != nil {
				panic(err)
			}

			systemBootstrappers = append(systemBootstrappers, networkingProviderSystemBootstrapper)
		}

		systemBootstrappers = append(systemBootstrappers, cloudProviderSystemBootstrapper)

		var systemResources *systembootstrapper.SystemResources
		if options.DryRun {
			systemResources = systembootstrap.GetBootstrapResources(clusterID, initialSystemID, initialSystemDefinitionURL, systemBootstrappers)
		} else {
			fmt.Printf("bootstrapping initial system \"%v\"\n", initialSystemIDString)
			systemResources, err = systembootstrap.Bootstrap(
				clusterID,
				initialSystemID,
				initialSystemDefinitionURL,
				systemBootstrappers,
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
	Cmd.Flags().StringVar(&options.MasterComponents.LatticeControllerManager.TerraformModulePath, "lattice-controller-manager-terraform-module-path", "", "optional path to terraform modules")
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

func parseCloudProviderVars() (*cloudprovider.ClusterBootstrapperOptions, *cloudprovider.SystemBootstrapperOptions, error) {
	var clusterOptions *cloudprovider.ClusterBootstrapperOptions
	var systemOptions *cloudprovider.SystemBootstrapperOptions

	switch cloudProviderName {
	case constants.ProviderLocal:
		clusterBootstrapperOptions, err := parseCloudProviderVarsLocal()
		if err != nil {
			return nil, nil, err
		}

		clusterOptions = &cloudprovider.ClusterBootstrapperOptions{
			Local: clusterBootstrapperOptions,
		}

		systemOptions = &cloudprovider.SystemBootstrapperOptions{
			Local: &localcloudprovider.SystemBootstrapperOptions{},
		}

	case constants.ProviderAWS:
		clusterBootstrapperOptions, err := parseProviderCloudVarsAWS()
		if err != nil {
			return nil, nil, err
		}

		clusterOptions = &cloudprovider.ClusterBootstrapperOptions{
			AWS: clusterBootstrapperOptions,
		}

		systemOptions = &cloudprovider.SystemBootstrapperOptions{
			AWS: &awscloudprovider.SystemBootstrapperOptions{},
		}

	default:
		return nil, nil, fmt.Errorf("unsupported cloudProviderName: %v", cloudProviderName)
	}

	return clusterOptions, systemOptions, nil
}

func parseCloudProviderVarsLocal() (*localcloudprovider.ClusterBootstrapperOptions, error) {
	options := &localcloudprovider.ClusterBootstrapperOptions{}
	flatStruct := localCloudOptionsFlat{}

	flags := cli.EmbeddedFlag{
		Target: &flatStruct,
		Expected: map[string]cli.EmbeddedFlagValue{
			"cluster-ip": {
				Required:     true,
				EncodingName: "IP",
			},
			"dns-controller-image": {
				Required:     true,
				EncodingName: "DNSControllerImage",
			},
			"dnsmasq-nanny-image": {
				Required:     true,
				EncodingName: "DnsmasqNannyImage",
			},
			"dnsmasq-nanny-args": {
				Required:     false,
				EncodingName: "DnsmasqNannyArgs",
				ValueParser: func(value string) (interface{}, error) {
					var argsWithoutPrefix = strings.Join(strings.Split(value, "=")[1:], "=")
					return strings.Split(argsWithoutPrefix, ":"), nil
				},
			},
			"dns-controller-args": {
				Required:     false,
				EncodingName: "DNSControllerArgs",
				ValueParser: func(value string) (interface{}, error) {
					var argsWithoutPrefix = strings.Join(strings.Split(value, "=")[1:], "=")
					return strings.Split(argsWithoutPrefix, ","), nil
				},
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}

	options.IP = flatStruct.IP
	options.DNS = &localcloudprovider.OptionsDNS{}
	options.DNS.ControllerArgs = flatStruct.DNSControllerArgs
	options.DNS.DnsmasqNannyArgs = flatStruct.DnsmasqNannyArgs
	options.DNS.DnsmasqNannyImage = flatStruct.DnsmasqNannyImage
	options.DNS.ControllerImage = flatStruct.DNSControllerImage

	return options, nil
}

func parseProviderCloudVarsAWS() (*awscloudprovider.ClusterBootstrapperOptions, error) {
	options := &awscloudprovider.ClusterBootstrapperOptions{}
	flags := cli.EmbeddedFlag{
		Target: &options,
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
	return options, nil
}

func parseServiceMeshVars() (*servicemesh.ClusterBootstrapperOptions, *servicemesh.SystemBootstrapperOptions, error) {
	var clusterOptions *servicemesh.ClusterBootstrapperOptions
	var systemOptions *servicemesh.SystemBootstrapperOptions

	switch serviceMeshProvider {
	case constants.ServiceMeshEnvoy:
		clusterBootstrapperOptions, err := parseServiceMeshVarsEnvoy()
		if err != nil {
			return nil, nil, err
		}

		clusterOptions = &servicemesh.ClusterBootstrapperOptions{
			Envoy: clusterBootstrapperOptions,
		}

		systemOptions = &servicemesh.SystemBootstrapperOptions{
			Envoy: &envoy.SystemBootstrapperOptions{
				XDSAPIImage: clusterBootstrapperOptions.XDSAPIImage,
			},
		}

	default:
		return nil, nil, fmt.Errorf("unsupported service mesh provider: %v", serviceMeshProvider)
	}

	return clusterOptions, systemOptions, nil
}

func parseServiceMeshVarsEnvoy() (*envoy.ClusterBootstrapperOptions, error) {
	envoyConfig := &envoy.ClusterBootstrapperOptions{}
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

func parseNetworkingVars() (*networkingprovider.ClusterBootstrapperOptions, *networkingprovider.SystemBootstrapperOptions, error) {
	var clusterOptions *networkingprovider.ClusterBootstrapperOptions
	var systemOptions *networkingprovider.SystemBootstrapperOptions

	switch networkingProviderName {
	case networkingprovider.Flannel:
		flannelOptions, err := parseNetworkingVarsFlannel()
		if err != nil {
			return nil, nil, err
		}

		clusterOptions = &networkingprovider.ClusterBootstrapperOptions{
			Flannel: flannelOptions,
		}

		systemOptions = &networkingprovider.SystemBootstrapperOptions{
			Flannel: &flannel.SystemBootstrapperOptions{},
		}

	case networkingprovider.None:
		noneOptions, err := parseNetworkingVarsNone()
		if err != nil {
			return nil, nil, err
		}

		clusterOptions = &networkingprovider.ClusterBootstrapperOptions{
			None: noneOptions,
		}

		systemOptions = &networkingprovider.SystemBootstrapperOptions{
			None: &none.SystemBootstrapperOptions{},
		}

	default:
		return nil, nil, fmt.Errorf("unsupported networking provider: %v", networkingProviderName)
	}

	return clusterOptions, systemOptions, nil
}

func parseNetworkingVarsFlannel() (*flannel.ClusterBootstrapperOptions, error) {
	flannelOptions := &flannel.ClusterBootstrapperOptions{}
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

func parseNetworkingVarsNone() (*none.ClusterBootstrapperOptions, error) {
	noneOptions := &none.ClusterBootstrapperOptions{}
	return noneOptions, nil
}
