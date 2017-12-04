package app

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/constants"
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/client"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
)

var (
	workingDir               string
	kubeconfigPath           string
	debug                    bool
	systemDefinitionUrl      string
	systemId                 string
	latticeContainerRegistry string
	componentBuildRegistry   string
	dockerAPIVersion         string
	provider                 string
	providerVars             *[]string
	terraformBackend         string
	terraformBackendVars     *[]string

	kubeConfig    *rest.Config
	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bootstrap-lattice",
	Short: "Bootstraps a kubernetes cluster to run lattice",
	Run: func(cmd *cobra.Command, args []string) {

		seedNamespaces()
		seedCrds()
		seedRbac()
		seedConfig(systemDefinitionUrl)
		seedEnvoyXdsApi()
		seedLatticeControllerManager()
		seedLatticeSystemEnvironmentManagerAPI()

		if provider == constants.ProviderLocal {
			seedLocalSpecific(systemId)
		} else {
			seedCloudSpecific()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().StringVar(&workingDir, "working-directory", "/tmp/lattice-system/", "path where subcommands will use as their working directory")
	RootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig-path", "", "path to kubeconfig to use if not being invoked from within kubernetes")
	RootCmd.Flags().BoolVar(&debug, "debug", false, "whether or not to use the debug version of container images")
	RootCmd.Flags().StringVar(&systemDefinitionUrl, "system-definition-url", "", "url of the system definition repo for the system")
	RootCmd.Flags().StringVar(&systemId, "system-id", "", "ID of the system")
	RootCmd.Flags().StringVar(&latticeContainerRegistry, "lattice-container-registry", "", "registry which stores the lattice infrastructure containers")
	RootCmd.Flags().StringVar(&componentBuildRegistry, "component-build-registry", "", "registry where component builds are tagged and potentially pushed to")
	RootCmd.Flags().StringVar(&dockerAPIVersion, "docker-api-version", "", "version of the docker API used by the docker daemons")
	RootCmd.Flags().StringVar(&provider, "provider", "", "provider")
	RootCmd.Flags().StringVar(&terraformBackend, "terraform-backend", "", "backend to use for storing terraform state")

	// Flags().StringArray --provider-var=a,b --provider-var=c results in ["a,b", "c"],
	// whereas Flags().StringSlice --provider-var=a,b --provider-var=c results in ["a", "b", "c"].
	// We don't want this because we want to be able to pass in for example
	// --provider-var=availability-zones=us-east-1a,us-east-1b resulting in ["availability-zones=us-east-1a,us-east-1b"]
	providerVars = RootCmd.Flags().StringArray("provider-var", nil, "additional variables to pass to the provider")
	terraformBackendVars = RootCmd.Flags().StringArray("terraform-backend-var", nil, "additional variables to pass to the terraform backend")
}

func initializeClients() {
	switch provider {
	case constants.ProviderLocal, constants.ProviderAWS:
	default:
		panic("unsupported provider")
	}

	var err error
	if kubeconfigPath == "" {
		kubeConfig, err = rest.InClusterConfig()
	} else {
		// TODO: support passing in the context when supported
		// https://github.com/kubernetes/minikube/issues/2100
		//configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			configOverrides,
		).ClientConfig()
	}

	if err != nil {
		panic(err)
	}

	kubeClient = kubeclientset.NewForConfigOrDie(kubeConfig)
	latticeClient = latticeclientset.NewForConfigOrDie(kubeConfig)
}
