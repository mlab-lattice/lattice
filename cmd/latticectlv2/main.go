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

var Latticectl = latticectl.Latticectl{
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
					&systems.StatusCommand{},
					&systems.DeleteCommand{},
					// Version commands
					&systems.ListVersionsCommand{},
					// Build commands
					&systems.BuildCommand{},
					&builds.ListBuildsCommand{
						Subcommands: []latticectl.Command{
							&builds.StatusCommand{},
						},
					},
					// Deploy commands
					&systems.DeployCommand{},
					&deploys.ListDeploysCommand{
						Subcommands: []latticectl.Command{
							&deploys.StatusCommand{},
						},
					},
					// Teardown commands
					&systems.TeardownCommand{},
					&teardowns.ListTeardownsCommand{
						Subcommands: []latticectl.Command{
							&teardowns.StatusCommand{},
						},
					},
					// Secret commands
					&secrets.ListSecretsCommand{
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

func main() {
	Latticectl.ExecuteColon()
}
