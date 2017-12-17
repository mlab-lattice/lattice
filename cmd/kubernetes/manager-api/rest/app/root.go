package app

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/managerapi/server/user/backend"
	"github.com/mlab-lattice/system/pkg/managerapi/server/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	kubeconfig       string
	clusterIDString  string
	port             int
	workingDirectory string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "manager-api",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID := types.ClusterID(clusterIDString)

		kubernetesBackend, err := backend.NewKubernetesBackend(clusterID, kubeconfig)
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
	goflag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	goflag.StringVar(&clusterIDString, "cluster-id", "", "id of the lattice cluster")
	goflag.StringVar(&workingDirectory, "workingDirectory", "/tmp/lattice-manager-api", "working directory to use")
	goflag.IntVar(&port, "port", 8080, "port to bind to")
}

func initCmd() {
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})
}
