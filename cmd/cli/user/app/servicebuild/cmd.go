package servicebuild

import (
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	asJSON          bool
	namespaceString string
	url             string
	namespace       types.LatticeNamespace
	userClient      user.Client
	namespaceClient cli.NamespaceClient
)

var Cmd = &cobra.Command{
	Use:  "service-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list service builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		namespaceClient.ServiceBuilds().List()
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get service build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ServiceBuildID(args[0])
		namespaceClient.ServiceBuilds().Show(id)
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(constants.UserSystemNamespace), "namespace to use")
	Cmd.PersistentFlags().BoolVarP(&asJSON, "json", "", false, "whether or not to display output as JSON")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}

func initCmd() {
	namespace = types.LatticeNamespace(namespaceString)

	userClient = rest.NewClient(url)
	restClient := userClient.Namespace(namespace)
	namespaceClient = cli.NewNamespaceClient(restClient, asJSON)
}
