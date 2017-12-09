package bootstrap

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper"
	baseboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/base"
	cloudboostrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/cloud"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/cli"

	"github.com/spf13/cobra"
)

var (
	kubeconfigPath      string
	kubeconfigContext   string
	kubeNamespacePrefix string

	latticeControllerManagerImage       string
	latticeControllerManagerArgs        []string
	defaultLatticeControllerManagerArgs = []string{
		"-v", "5",
		"-logtostderr",
	}

	managerAPIImage       string
	managerAPIPort        int32
	managerAPIHostNetwork bool
	managerAPIArgs        []string
	defaultManagerAPIArgs = []string{}

	componentBuilderImage            string
	componentBuilderDockerAPIVersion string

	componentBuildDockerArtifactRegistry           string
	componentBuildDockerArtifactRepositoryPerImage bool
	componentBuildDockerArtifactRepository         string
	componentBuildDockerArtifactPush               bool

	envoyPrepareImage      string
	envoyImage             string
	envoyRedirectCIDRBlock string
	envoyXDSAPIPort        int32

	provider     string
	providerVars []string

	terraformBackend     string
	terraformBackendVars []string
)

var Cmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "bootstraps a kubernetes cluster to run Lattice",
	// FIXME: figure out why it thinks two args are getting passed in
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !componentBuildDockerArtifactRepositoryPerImage && componentBuildDockerArtifactRepository == "" {
			panic("must specify component-build-docker-artifact-repository if not component-build-docker-artifact-repository-per-image")
		}

		kubeconfig, err := kubeutil.NewConfig(kubeconfigPath, "")
		if err != nil {
			panic(err)
		}

		providerConfig, err := parseProviderVars()
		if err != nil {
			panic(err)
		}

		terraformConfig, err := parseTerraformVars()
		if err != nil {
			panic(err)
		}

		options := &bootstrapper.Options{
			Config: crv1.ConfigSpec{
				KubernetesNamespacePrefix: kubeNamespacePrefix,
				ComponentBuild: crv1.ConfigComponentBuild{
					Builder: crv1.ConfigComponentBuildBuilder{
						Image:            componentBuilderImage,
						DockerAPIVersion: componentBuilderDockerAPIVersion,
					},
					DockerArtifact: crv1.ConfigComponentBuildDockerArtifact{
						Registry:           componentBuildDockerArtifactRegistry,
						RepositoryPerImage: componentBuildDockerArtifactRepositoryPerImage,
						Repository:         componentBuildDockerArtifactRepository,
						Push:               componentBuildDockerArtifactPush,
					},
				},
				Envoy: crv1.ConfigEnvoy{
					PrepareImage:      envoyPrepareImage,
					Image:             envoyImage,
					RedirectCIDRBlock: envoyRedirectCIDRBlock,
					XDSAPIPort:        envoyXDSAPIPort,
				},
				SystemConfigs: map[types.LatticeNamespace]crv1.ConfigSystem{},
				Provider:      *providerConfig,
				Terraform:     terraformConfig,
			},
			MasterComponents: baseboostrapper.MasterComponentOptions{
				LatticeControllerManager: baseboostrapper.LatticeControllerManagerOptions{
					Image: latticeControllerManagerImage,
					Args:  latticeControllerManagerArgs,
				},
				ManagerAPI: baseboostrapper.ManagerAPIOptions{
					Image:       managerAPIImage,
					Port:        managerAPIPort,
					HostNetwork: managerAPIHostNetwork,
					Args:        managerAPIArgs,
				},
			},
			Networking: &cloudboostrapper.NetworkingOptions{
				Flannel: &cloudboostrapper.FlannelOptions{},
			},
		}

		b, err := bootstrapper.NewBootstrapper(options, kubeconfig)
		if err != nil {
			panic(err)
		}

		if err := b.Bootstrap(); err != nil {
			panic(err)
		}
	},
}

func init() {
	Cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig")
	Cmd.Flags().StringVar(&kubeconfigContext, "kubeconfig-context", "", "context in the kubeconfig to use")

	Cmd.Flags().StringVar(&kubeNamespacePrefix, "namespace-prefix", "lattice", "prefix to add to namespaces that lattice will create and own")

	Cmd.Flags().StringVar(&latticeControllerManagerImage, "lattice-controller-manager-image", "", "docker image to user for the lattice-controller-manager")
	Cmd.MarkFlagRequired("lattice-controller-manager-image")
	Cmd.Flags().StringArrayVar(&latticeControllerManagerArgs, "lattice-controller-manager-args", defaultLatticeControllerManagerArgs, "extra arguments (besides --provider) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&managerAPIImage, "manager-api-image", "", "docker image to user for the lattice-controller-manager")
	Cmd.MarkFlagRequired("manager-api-image")
	Cmd.Flags().Int32Var(&managerAPIPort, "manager-api-port", 80, "port that the manager-api should listen on")
	Cmd.Flags().BoolVar(&managerAPIHostNetwork, "manager-api-host-network", true, "whether or not the manager-api should be on the host network")
	Cmd.Flags().StringArrayVar(&managerAPIArgs, "manager-api-args", defaultManagerAPIArgs, "extra arguments (besides --provider) to pass to the lattice-controller-manager")

	Cmd.Flags().StringVar(&componentBuilderImage, "component-builder-image", "", "docker image to user for the component-builder")
	Cmd.MarkFlagRequired("component-builder-image")
	Cmd.Flags().StringVar(&componentBuilderDockerAPIVersion, "component-builder-docker-api-version", "", "version of the docker API used by the build node docker daemon")

	Cmd.Flags().StringVar(&componentBuildDockerArtifactRegistry, "component-build-docker-artifact-registry", "", "registry to tag component build docker artifacts with")
	Cmd.MarkFlagRequired("component-build-docker-artifact-registry")
	Cmd.Flags().BoolVar(&componentBuildDockerArtifactRepositoryPerImage, "component-build-docker-artifact-repository-per-image", false, "if false, one repository with a new tag for each artifact will be use, if true a new repository for each artifact will be used")
	Cmd.Flags().StringVar(&componentBuildDockerArtifactRepository, "component-build-docker-artifact-repository", "", "repository to tag component build docker artifacts with, required if component-build-docker-artifact-repository-per-image is false")
	Cmd.Flags().BoolVar(&componentBuildDockerArtifactPush, "component-build-docker-artifact-push", true, "whether or not the component-builder should push the docker artifact (use false for local)")

	Cmd.Flags().StringVar(&envoyPrepareImage, "envoy-prepare-image", "", "image to use for envoy-prepare")
	Cmd.MarkFlagRequired("envoy-prepare-image")
	Cmd.Flags().StringVar(&envoyImage, "envoy-image", "envoyproxy/envoy-alpine", "image to use for envoy")
	Cmd.Flags().StringVar(&envoyRedirectCIDRBlock, "envoy-redirect-cidr-block", "", "CIDR block to use to redirect traffic to envoy")
	Cmd.MarkFlagRequired("envoy-redirect-cidr-block")
	Cmd.Flags().Int32Var(&envoyXDSAPIPort, "envoy-xds-api-port", 8080, "port that the envoy-xds-api should listen on")

	Cmd.Flags().StringVar(&provider, "provider", "", "provider that the cluster is being bootstrapped on")
	Cmd.Flags().StringArrayVar(&providerVars, "provider-var", nil, "additional variables for the provider")

	Cmd.Flags().StringVar(&terraformBackend, "terraform-backend", "", "backend to use for terraform")
	Cmd.Flags().StringArrayVar(&terraformBackendVars, "terraform-backend-var", nil, "additional variables for the terraform backend")
}

func parseProviderVars() (*crv1.ConfigProvider, error) {
	var config *crv1.ConfigProvider
	switch provider {
	case constants.ProviderLocal:
		localConfig, err := parseProviderVarsLocal()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigProvider{
			Local: localConfig,
		}
	case constants.ProviderAWS:
		awsConfig, err := parseProviderVarsAWS()
		if err != nil {
			return nil, err
		}
		config = &crv1.ConfigProvider{
			AWS: awsConfig,
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %v", provider)
	}

	return config, nil
}

func parseProviderVarsLocal() (*crv1.ConfigProviderLocal, error) {
	localConfig := &crv1.ConfigProviderLocal{}
	flags := cli.EmbeddedFlag{
		Target: &localConfig,
		Expected: map[string]cli.EmbeddedFlagValue{
			"system-ip": {
				Required:     true,
				EncodingName: "ip",
			},
		},
	}

	err := flags.Parse(providerVars)
	if err != nil {
		return nil, err
	}
	return localConfig, nil
}

func parseProviderVarsAWS() (*crv1.ConfigProviderAWS, error) {
	awsConfig := &crv1.ConfigProviderAWS{}
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

	err := flags.Parse(providerVars)
	if err != nil {
		return nil, err
	}
	return awsConfig, nil
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

	err := flags.Parse(providerVars)
	if err != nil {
		return nil, err
	}
	return s3Config, nil
}
