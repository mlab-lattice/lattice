package app

import (
	"github.com/mlab-lattice/system/cmd/latticectlv2/app/system"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

//// Cmd represents the base command when called without any subcommands
//var Cmd = &cobra.BasicCommand{
//	Use:   "latticectl",
//	Short: "BasicCommand line utility for interacting with lattice clusters and systems",
//}

//// Execute adds all child commands to the root command and sets flags appropriately.
//// This is called by main.main(). It only needs to happen once to the rootCmd.
//func Execute() {
//	if err := Cmd.Execute(); err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//}

//func init() {
//	Cmd.AddCommand(cluster.Cmd)
//	Cmd.AddCommand(definition.Cmd)
//	Cmd.AddCommand(system.Cmd)
//}

var Cmd = command.BasicCommand{
	Name:  "latticectl",
	Short: "BasicCommand line utility for interacting with lattice clusters and systems",
	Subcommands: []command.Command{
		system.Cmd,
	},
}
