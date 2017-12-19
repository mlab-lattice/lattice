package systembuild

import (
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	output          string
	namespaceString string
	url             string
	namespace       types.LatticeNamespace
	userClient      user.Client
	namespaceClient user.NamespaceClient
)

var Cmd = &cobra.Command{
	Use:  "system-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list system builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := namespaceClient.SystemBuilds()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowSystemBuilds(builds, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get system build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.SystemBuildID(args[0])
		build, err := namespaceClient.SystemBuild(id).Get()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowSystemBuild(build, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(constants.UserSystemNamespace), "namespace to use")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}

func initCmd() {
	namespace = types.LatticeNamespace(namespaceString)

	userClient = rest.NewClient(url)
	namespaceClient = userClient.Namespace(namespace)
}
