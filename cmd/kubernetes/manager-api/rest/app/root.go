package app

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/managerapi/server/backend"
	"github.com/mlab-lattice/system/pkg/managerapi/server/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	kubeconfig       string
	clusterIDString  string
	port             int
	workingDirectory string

	cloudProviderName string
	cloudProviderVars []string

	serviceMeshProvider     string
	serviceMeshProviderVars []string

	networkingProviderName string
	networkingProviderVars []string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "manager-api",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID := types.ClusterID(clusterIDString)

		cloudSystemBootstrapper, err := cloudprovider.SystemBootstrapperFromFlags(cloudProviderName, cloudProviderVars)
		if err != nil {
			panic(err)
		}

		serviceMeshSystemBootstrapper, err := servicemesh.SystemBootstrapperFromFlags(serviceMeshProvider, serviceMeshProviderVars)
		if err != nil {
			panic(err)
		}

		networkingProviderSystemBoostrapper, err := networkingprovider.SystemBootstrapperFromFlags(networkingProviderName, networkingProviderVars)
		if err != nil {
			panic(err)
		}

		systemBoostrappers := []bootstrapper.Interface{
			serviceMeshSystemBootstrapper,
			networkingProviderSystemBoostrapper,
			cloudSystemBootstrapper,
		}

		kubernetesBackend, err := backend.NewKubernetesBackend(
			clusterID,
			kubeconfig,
			systemBoostrappers,
		)
		if err != nil {
			panic(err)
		}

		rest.RunNewRestServer(kubernetesBackend, int32(port), workingDirectory)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(initCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)

	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	RootCmd.Flags().StringVar(&clusterIDString, "cluster-id", "", "id of the lattice cluster")
	RootCmd.Flags().StringVar(&workingDirectory, "workingDirectory", "/tmp/lattice-manager-api", "working directory to use")
	RootCmd.Flags().IntVar(&port, "port", 8080, "port to bind to")

	RootCmd.Flags().StringVar(&cloudProviderName, "cloud-provider", "", "cloud provider that the cluster is being bootstrapped on")
	RootCmd.MarkFlagRequired("cloud-provider")
	RootCmd.Flags().StringArrayVar(&cloudProviderVars, "cloud-provider-var", nil, "additional variables for the cloud provider")

	RootCmd.Flags().StringVar(&serviceMeshProvider, "service-mesh", "", "service mesh provider to use")
	RootCmd.MarkFlagRequired("service-mesh")
	RootCmd.Flags().StringArrayVar(&serviceMeshProviderVars, "service-mesh-var", nil, "additional variables for the cloud provider")

	RootCmd.Flags().StringVar(&networkingProviderName, "networking-provider", "", "provider to use for networking")
	RootCmd.Flags().StringArrayVar(&networkingProviderVars, "networking-provider-var", nil, "additional variables for the networking provider")
}

func initCmd() {
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})
}
