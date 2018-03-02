package app

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/context"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems/deploys"
)

var cmd = &latticectl.BaseCommand{
	Name:    "latticectl",
	Short:   "command line utility for interacting with lattice clusters and systems",
	Client:  latticectl.DefaultLatticeClient,
	Context: &latticectl.DefaultFileContext{},
	Subcommands: []command.Command2{
		&context.Command{
			Subcommands: []command.Command2{
				&context.GetCommand{},
			},
		},
		&systems.Command{
			Subcommands: []command.Command2{
				&systems.CreateCommand{},
				&systems.GetCommand{},
				&systems.DeleteCommand{},
				&systems.BuildCommand{},
				&systems.DeployCommand{},
				&deploys.Command{
					Subcommands: []command.Command2{
						&deploys.GetCommand{},
					},
				},
			},
		},
	},
}

func Execute() {
	//command.Execute(cmd)
	command.ExecuteColon(cmd)
}
