package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/provisioner"

	"github.com/spf13/cobra"
)

var (
	backend     string
	backendVars []string
)

var provisionSystemCmd = &cobra.Command{
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
			config, err := parseBackendKubernetesVars(provider)
			if err != nil {
				panic(fmt.Sprintf("error parsing kubernetes backend vars: %v", err))
			}

			provisioner, err = getKubernetesProvisioner(provider, name, actionProvision, config)
			if err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("unsupported backend %v", backend))
		}

		err := provisioner.Provision(name, url)
		if err != nil {
			panic(err)
		}

		addr, err := provisioner.Address(name)
		if err != nil {
			panic(err)
		}

		fmt.Printf("System Environment Manager address:\n%v\n", addr)
	},
}

func init() {
	RootCmd.AddCommand(provisionSystemCmd)

	provisionSystemCmd.Flags().StringVar(&backend, "backend", constants.BackendTypeKubernetes, "lattice backend to use")
	provisionSystemCmd.Flags().StringArrayVar(&backendVars, "backend-var", nil, "additional variables to pass in to the backend")
}
