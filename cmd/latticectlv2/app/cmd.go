package app

import (
	"github.com/mlab-lattice/system/cmd/latticectlv2/app/systems"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

var Cmd = command.BaseCommand{
	Name:  "latticectl",
	Short: "command line utility for interacting with lattice clusters and systems",
	Subcommands: []command.Command{
		systems.Cmd,
	},
}
