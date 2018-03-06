package main

import (
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/context"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/local"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/commands/systems/deploys"
)

func main() {
	lctl := latticectl.Latticectl{
		Client:  latticectl.DefaultLatticeClient,
		Context: &latticectl.DefaultFileContext{},
		Root: &latticectl.BaseCommand{
			Name:  "latticectl",
			Short: "command line utility for interacting with lattice clusters and systems",
			Subcommands: []latticectl.Command{
				&context.Command{
					Subcommands: []latticectl.Command{
						&context.GetCommand{},
						&context.SetCommand{},
					},
				},
				&local.Command{
					Subcommands: []latticectl.Command{
						&local.DownCommand{},
						&local.UpCommand{},
					},
				},
				&systems.Command{
					Subcommands: []latticectl.Command{
						&systems.CreateCommand{},
						&systems.GetCommand{},
						&systems.DeleteCommand{},
						&systems.BuildCommand{},
						&systems.DeployCommand{},
						&deploys.Command{
							Subcommands: []latticectl.Command{
								&deploys.GetCommand{},
							},
						},
					},
				},
			},
		},
	}

	//lctl.Execute()
	lctl.ExecuteColon()
}
