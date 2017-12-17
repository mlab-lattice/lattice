package componentbuild

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	follow bool

	systemIDString string
	url            string
	systemID       types.SystemID
	userClient     user.Client
	systemClient   user.SystemClient
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
		builds, err := systemClient.ComponentBuilds()
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
	Short: "get component build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		build, err := systemClient.ComponentBuild(id).Get()
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

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "get component build logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		logs, err := systemClient.ComponentBuild(id).Logs(follow)
		if err != nil {
			log.Fatal(err)
		}
		defer logs.Close()

		io.Copy(os.Stdout, logs)
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&systemIDString, "system", string(constants.SystemIDDefault), "system to use")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "whether or not to follow the logs")
}

func initCmd() {
	systemID = types.SystemID(systemIDString)

	userClient = rest.NewClient(url)
	systemClient = userClient.System(systemID)
}
