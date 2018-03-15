package main

import (
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/context"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/local"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/services"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/builds"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/deploys"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/secrets"
	"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems/teardowns"
)

func main() {
	lctl := latticectl.Latticectl{
		Client:  latticectl.DefaultLatticeClient,
		Context: &latticectl.DefaultFileContext{},
		Root: &latticectl.BaseCommand{
			Name:  "latticectl",
			Short: "command line utility for interacting with lattice clusters and systems",
			Subcommands: []latticectl.Command{
				// Context commands
				&context.Command{
					Subcommands: []latticectl.Command{
						&context.GetCommand{},
						&context.SetCommand{},
					},
				},
				// Local commands
				&local.Command{
					Subcommands: []latticectl.Command{
						&local.DownCommand{},
						&local.UpCommand{},
					},
				},
				// System commands
				&systems.ListSystemsCommand{
					Subcommands: []latticectl.Command{
						&systems.CreateCommand{},
						&systems.GetCommand{},
						&systems.DeleteCommand{},
						&systems.BuildCommand{},
						// Build commands
						&builds.ListBuildsCommand{
							Subcommands: []latticectl.Command{
								&builds.GetCommand{},
							},
						},
						// Deploy commands
						&systems.DeployCommand{},
						&deploys.Command{
							Subcommands: []latticectl.Command{
								&deploys.GetCommand{},
							},
						},
						// Teardown commands
						&systems.TeardownCommand{},
						&teardowns.Command{
							Subcommands: []latticectl.Command{
								&teardowns.GetCommand{},
							},
						},
						// Secret commands
						&secrets.Command{
							Subcommands: []latticectl.Command{
								&secrets.GetCommand{},
								&secrets.SetCommand{},
							},
						},
					},
				},
				// Service commands
				&services.ListServicesCommand{
					Subcommands: []latticectl.Command{
						&services.StatusCommand{},
						&services.AddressCommand{},
					},
				},
			},
		},
	}

	//lctl.Execute()
	lctl.ExecuteColon()
}
