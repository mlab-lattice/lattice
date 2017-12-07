package app

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	nodeID                    string
	masterComponentLogsFollow bool
)

var masterCmd = &cobra.Command{
	Use:  "master",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var masterComponentsCommand = &cobra.Command{
	Use:  "components",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var masterComponentsListCommand = &cobra.Command{
	Use:  "list",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		components, err := adminClient.Master().Components()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(components)
	},
}

var masterComponentsLogsCommand = &cobra.Command{
	Use:  "logs [component]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		component := args[0]
		logs, err := adminClient.Master().Component(component).Logs(nodeID, masterComponentLogsFollow)
		if err != nil {
			log.Fatal(err)
		}
		defer logs.Close()

		io.Copy(os.Stdout, logs)
	},
}

var masterComponentsRestartCommand = &cobra.Command{
	Use:  "restart [component]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		component := args[0]
		err := adminClient.Master().Component(component).Restart(nodeID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(masterCmd)

	masterCmd.PersistentFlags().StringVar(&nodeID, "node", "0", "master node to query")
	masterCmd.AddCommand(masterComponentsCommand)

	masterComponentsCommand.AddCommand(masterComponentsListCommand)

	masterComponentsCommand.AddCommand(masterComponentsLogsCommand)
	masterComponentsLogsCommand.Flags().BoolVarP(&masterComponentLogsFollow, "follow", "f", false, "whether or not to follow the logs")

	masterComponentsCommand.AddCommand(masterComponentsRestartCommand)
}
