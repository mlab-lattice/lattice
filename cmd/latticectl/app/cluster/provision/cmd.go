package provision

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"

	"github.com/spf13/cobra"
)

var (
	workDir     string
	backend     string
	backendVars []string
)

var Cmd = &cobra.Command{
	Use:   "provision [PROVIDER] [NAME] [URL]",
	Short: "Provision a system",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		name := args[1]
		url := args[2]

		var provisioner provisioner.Interface
		switch backend {
		case constants.BackendTypeKubernetes:
			var err error
			provisioner, err = getKubernetesProvisioner(provider)
			if err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("unsupported backend %v", backend))
		}

		clusterAddress, err := provisioner.Provision(name, url)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Cluster Manager address:\n%v\n", clusterAddress)
	},
}

func init() {
	Cmd.Flags().StringVar(&workDir, "work-directory", "/tmp/lattice/cluster", "path where subcommands will use as their working directory")
	Cmd.Flags().StringVar(&backend, "backend", constants.BackendTypeKubernetes, "lattice backend to use")
	Cmd.Flags().StringArrayVar(&backendVars, "backend-var", nil, "additional variables to pass in to the backend")
}
