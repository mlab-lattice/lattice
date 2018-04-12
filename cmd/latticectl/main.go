package main

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/context"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/kubernetes/bootstrap"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/local"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/services"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/builds"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/deploys"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/secrets"
	"github.com/mlab-lattice/lattice/pkg/latticectl/commands/systems/teardowns"
)

var Latticectl = latticectl.Latticectl{
	Client:  latticectl.DefaultLatticeClient,
	Context: &latticectl.DefaultFileContext{},
	Root: &latticectl.BaseCommand{
		Name:  "latticectl",
		Short: "command line utility for interacting with lattices and systems",
		Subcommands: []latticectl.Command{
			// Context commands
			&context.Command{
				Subcommands: []latticectl.Command{
					&context.SetCommand{},
				},
			},
			// Kubernetes commands
			&kubernetes.Command{
				Subcommands: []latticectl.Command{
					&bootstrap.Command{},
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
