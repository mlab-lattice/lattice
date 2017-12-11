package servicebuild

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	namespaceString string
	url             string
	namespace       types.LatticeNamespace
	userClient      user.Client
	namespaceClient user.NamespaceClient
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
		builds, err := namespaceClient.ServiceBuilds()
		if err != nil {
			log.Fatal(err)
		}

		buf, err := json.MarshalIndent(builds, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get service build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ServiceBuildID(args[0])
		build, err := namespaceClient.ServiceBuild(id).Get()
		if err != nil {
			log.Fatal(err)
		}

		buf, err := json.MarshalIndent(build, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(constants.NamespaceDefault), "namespace to use")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}

func initCmd() {
	namespace = types.LatticeNamespace(namespaceString)

	userClient = rest.NewClient(url)
	namespaceClient = userClient.Namespace(namespace)
}
