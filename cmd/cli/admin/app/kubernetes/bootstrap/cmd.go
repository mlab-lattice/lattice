package bootstrap

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper"
	basebootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/base"
	cloudbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/cloud"
	localbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/local"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/cli"

	"k8s.io/apimachinery/pkg/api/errors"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	kubeconfigPath    string
	kubeconfigContext string

	defaultLatticeControllerManagerArgs = []string{
		"-v", "5",
		"--logtostderr",
	}

	defaultLocalDNSControllerArgs = []string {
		"-v", "5",
		"--logtostderr",
		"--resolv", "/etc/k8s/dns/dnsmasq-nanny/resolv.conf",
		"--extraconf", "/etc/k8s/dns/dnsmasq-nanny/dnsmasq.conf",
	}

	defaultLocalDNSServerArgs = []string {
		// TODO :: Clean up - split into dnsmasq args and dnsnanny args.
		"-v=2",
		"-logtostderr",
		"-restartDnsmasq=true",
		"-configDir=/etc/k8s/dns/dnsmasq-nanny",
		// Arguments after -- are passed straight to dnsmasq.
		"--",
		"-k", //Keep in foreground so as to not immediately exit.
	}

	defaultManagerAPIArgs = []string{}

	clusterIDString string

	initialSystemIDString      string
	initialSystemDefinitionURL string

	cloudProvider     string
	cloudProviderVars []string

	serviceMeshProvider     string
	serviceMeshProviderVars []string

	terraformBackend     string
	terraformBackendVars []string

	networkingProvider     string
	networkingProviderVars []string
)

var options = &bootstrapper.Options{
	Config: crv1.ConfigSpec{
		ComponentBuild: crv1.ConfigComponentBuild{
			Builder:        crv1.ConfigComponentBuildBuilder{},
			DockerArtifact: crv1.ConfigComponentBuildDockerArtifact{},
		},
		ServiceMesh: crv1.ConfigServiceMesh{},
	},
	MasterComponents: basebootstrapper.MasterComponentOptions{
		LatticeControllerManager: basebootstrapper.LatticeControllerManagerOptions{},
		ManagerAPI:               basebootstrapper.ManagerAPIOptions{},
	},
	LocalComponents: localbootstrapper.LocalComponentOptions{
		LocalDNSController: localbootstrapper.LocalDNSControllerOptions{},
		LocalDNSServer:	    localbootstrapper.LocalDNSServerOptions{},
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

		var kubeconfig *rest.Config
		if !options.DryRun {
			var err error
			kubeconfig, err = kubeutil.NewConfig(kubeconfigPath, "")
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

		networkingOptions, err := parseNetworkingVars()
		if err != nil {
			panic(err)
		}
		options.Networking = networkingOptions

		b, err := bootstrapper.NewBootstrapper(clusterID, options, kubeconfig)
		if err != nil {
			panic(err)
		}

		objects, err := b.Bootstrap()
		if err != nil {
			panic(err)
		}

		if options.DryRun {
			if initialSystemDefinitionURL != "" {
				resources := kubeutil.NewSystem(clusterID, initialSystemID, initialSystemDefinitionURL)
				objects = append(objects, []interface{}{resources.System, resources.Namespace}...)

				for _, sa := range resources.ServiceAccounts {
					objects = append(objects, interface{}(sa))
				}

				for _, roleBinding := range resources.RoleBindings {
					objects = append(objects, interface{}(roleBinding))
				}
			}

			output := ""
			for _, object := range objects {
				output += "---\n"
				data, err := yaml.Marshal(object)
				if err != nil {
					panic(err)
				}
				output += string(data)
			}
			fmt.Printf(output)
			return
		}

		if initialSystemDefinitionURL != "" {
			fmt.Printf("Seeding initial system \"%v\"\n", initialSystemIDString)
			kubeClient := kubeclientset.NewForConfigOrDie(kubeconfig)
			latticeClient := latticeclientset.NewForConfigOrDie(kubeconfig)
			_, err := kubeutil.CreateNewSystem(
				clusterID,
				initialSystemID,
				initialSystemDefinitionURL,
				kubeClient,
				latticeClient,
			)
			if err != nil && !errors.IsAlreadyExists(err) {
				panic(err)
			}
		}
	},
}

func init() {
	Cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "if set, will not actually bootstrap the cluster but will instead print out the resources needed to bootstrap the cluster")
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
	Cmd.Flags().StringArrayVar(&options.MasterComponents.LatticeControllerManager.Args, "lattice-controller-manager-args", defaultLatticeControllerManagerArgs, "extra arguments (besides --cloudProvider) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&options.MasterComponents.ManagerAPI.Image, "manager-api-image", "", "docker image to user for the lattice-controller-manager")
	Cmd.MarkFlagRequired("manager-api-image")
	Cmd.Flags().Int32Var(&options.MasterComponents.ManagerAPI.Port, "manager-api-port", 80, "port that the manager-api should listen on")
	Cmd.Flags().BoolVar(&options.MasterComponents.ManagerAPI.HostNetwork, "manager-api-host-network", true, "whether or not the manager-api should be on the host network")
	Cmd.Flags().StringArrayVar(&options.MasterComponents.ManagerAPI.Args, "manager-api-args", defaultManagerAPIArgs, "extra arguments (besides --cloudProvider) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&initialSystemIDString, "initial-system-name", "default", "name to use for the initial system if --initial-system-definition-url is set")
	Cmd.Flags().StringVar(&initialSystemDefinitionURL, "initial-system-definition-url", "", "URL to use for the definition of the optional initial system")

	Cmd.Flags().StringVar(&cloudProvider, "cloud-provider", "", "cloud provider that the cluster is being bootstrapped on")
	Cmd.MarkFlagRequired("cloud-provider")
	Cmd.Flags().StringArrayVar(&cloudProviderVars, "cloud-provider-var", nil, "additional variables for the cloud provider")

	Cmd.Flags().StringVar(&serviceMeshProvider, "service-mesh", "", "service mesh provider to use")
	Cmd.MarkFlagRequired("service-provider")
	Cmd.Flags().StringArrayVar(&serviceMeshProviderVars, "service-mesh-var", nil, "additional variables for the cloud provider")

	Cmd.Flags().StringVar(&options.LocalComponents.LocalDNSController.Image, "local-dns-controller-image", "", "docker image to use for the local-dns controller")
	Cmd.MarkFlagRequired("local-dns-controller-image")
	Cmd.Flags().StringArrayVar(&options.LocalComponents.LocalDNSController.Args, "local-dns-controller-args", defaultLocalDNSControllerArgs, "extra arguments (besides --provider) to pass to the local-dns-controller")

	Cmd.Flags().StringVar(&options.LocalComponents.LocalDNSServer.Image, "local-dns-server-image", "", "docker image to use for the local DNS server")
	Cmd.MarkFlagRequired("local-dns-server-image")
	Cmd.Flags().StringArrayVar(&options.LocalComponents.LocalDNSServer.Args, "local-dns-server-args", defaultLocalDNSServerArgs, "extra arguments to pass to the local-dns-server")

	Cmd.Flags().StringVar(&terraformBackend, "terraform-backend", "", "backend to use for terraform")
	Cmd.Flags().StringArrayVar(&terraformBackendVars, "terraform-backend-var", nil, "additional variables for the terraform backend")

	Cmd.Flags().StringVar(&networkingProvider, "networking-provider", "", "provider to use for networking")
	Cmd.Flags().StringArrayVar(&networkingProviderVars, "networking-provider-var", nil, "additional variables for the networking provider")
}

func parseCloudProviderVars() (*crv1.ConfigCloudProvider, error) {
	var config *crv1.ConfigCloudProvider
	switch cloudProvider {
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
		return nil, fmt.Errorf("unsupported cloudProvider: %v", cloudProvider)
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

func parseNetworkingVars() (*cloudbootstrapper.NetworkingOptions, error) {
	if networkingProvider == "" {
		return nil, nil
	}

	var options *cloudbootstrapper.NetworkingOptions
	switch terraformBackend {
	case kubeconstants.NetworkingProviderFlannel:
		flannelOptions, err := parseNetworkingVarsFlannel()
		if err != nil {
			return nil, err
		}
		options = &cloudbootstrapper.NetworkingOptions{
			Flannel: flannelOptions,
		}
	default:
		return nil, fmt.Errorf("unsupported networking provider: %v", networkingProvider)
	}

	return options, nil
}

func parseNetworkingVarsFlannel() (*cloudbootstrapper.FlannelOptions, error) {
	flannelOptions := &cloudbootstrapper.FlannelOptions{}
	flags := cli.EmbeddedFlag{
		Target: &flannelOptions,
		Expected: map[string]cli.EmbeddedFlagValue{
			"network-cidr-block": {
				Required:     true,
				EncodingName: "NetworkCIDRBlock",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return flannelOptions, nil
}
