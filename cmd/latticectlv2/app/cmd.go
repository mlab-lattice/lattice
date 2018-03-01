package app

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems/deploys"
)

var Cmd = command.BaseCommand{
	Name:  "latticectl",
	Short: "command line utility for interacting with lattice clusters and systems",
	Subcommands: []command.Command{
		&systems.Command{
			Client:  latticectl.DefaultLatticeClient,
			Context: &latticectl.DefaultFileContext{},
			Subcommands: []latticectl.LatticeCommand{
				&systems.CreateCommand{},
				&systems.GetCommand{},
				&systems.DeleteCommand{},
				&systems.BuildCommand{},
				&systems.DeployCommand{},
				&deploys.Command{
					Subcommands: []latticectl.LatticeCommand{
						&deploys.GetCommand{},
					},
				},
			},
		},
	},
}
