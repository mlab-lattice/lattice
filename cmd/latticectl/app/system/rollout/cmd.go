package rollout

import (
	"fmt"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	output string

	systemIDString string
	url            string
	systemID       types.SystemID
	userClient     client.Interface
	systemClient   client.SystemClient

	rolloutBuildID string
	rolloutVersion string
)

var Cmd = &cobra.Command{
	Use:  "rollout",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list rollouts",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		rollouts, err := systemClient.Rollouts(systemID).List()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowRollouts(rollouts, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get rollout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.DeployID(args[0])
		build, err := systemClient.Rollouts(systemID).Get(id)
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowRollout(build, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create rollout",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if rolloutBuildID == "" && rolloutVersion == "" {
			fmt.Println("must supply either build ID or version")
			os.Exit(1)
		}

		var rolloutID types.DeployID
		var err error
		if rolloutBuildID != "" {
			rolloutID, err = systemClient.Rollouts(systemID).CreateFromBuild(types.BuildID(rolloutBuildID))
		} else {
			rolloutID, err = systemClient.Rollouts(systemID).CreateFromVersion(rolloutVersion)
		}

		if err != nil {
			fmt.Printf("Received error creating rollout: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Sucessfully created rollout %v\n", rolloutID)
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&systemIDString, "system", string(constants.SystemIDDefault), "system to use")

	Cmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	Cmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	Cmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&rolloutBuildID, "build-id", "", "build ID to use for the rollout")
	createCmd.Flags().StringVar(&rolloutVersion, "version", "", "version to use for the rollout")
}

func initCmd() {
	systemID = types.SystemID(systemIDString)

	userClient = rest.NewClient(url)
	systemClient = userClient.Systems()
}
