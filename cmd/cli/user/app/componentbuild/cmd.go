package componentbuild

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
	follow bool
	output string

	namespaceString string
	url             string
	namespace       types.LatticeNamespace
	userClient      user.Client
	namespaceClient user.NamespaceClient
)

var Cmd = &cobra.Command{
	Use:  "component-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list component builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := namespaceClient.ComponentBuilds()
		if err != nil {
			log.Panic(err)
		}

		format := cli.GetFormatFromString(output)
		cli.ShowComponentBuilds(builds, format)
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get component build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		build, err := namespaceClient.ComponentBuild(id).Get()
		if err != nil {
			log.Panic(err)
		}

		format := cli.GetFormatFromString(output)
		cli.ShowComponentBuild(build, format)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "get component build logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		logs, err := namespaceClient.ComponentBuild(id).Logs(follow)
		if err != nil {
			log.Fatal(err)
		}
		cli.ShowComponentBuildLog(logs)
		logs.Close()
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(constants.UserSystemNamespace), "namespace to use")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "whether or not to follow the logs")
	getCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
}

func initCmd() {
	namespace = types.LatticeNamespace(namespaceString)

	userClient = rest.NewClient(url)
	namespaceClient = userClient.Namespace(namespace)
}
