package main

import (
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command/bootstrap"
	"github.com/mlab-lattice/system/pkg/latticectl/command/bootstrap/kubernetes"
	"github.com/mlab-lattice/system/pkg/latticectl/command/context"
	"github.com/mlab-lattice/system/pkg/latticectl/command/local"
	"github.com/mlab-lattice/system/pkg/latticectl/command/systems"
	"github.com/mlab-lattice/system/pkg/latticectl/command/systems/deploys"
	"github.com/mlab-lattice/system/pkg/latticectl/command/systems/secrets"
)

func main() {
	lctl := latticectl.Latticectl{
		Client:  latticectl.DefaultLatticeClient,
		Context: &latticectl.DefaultFileContext{},
		Root: &latticectl.BaseCommand{
			Name:  "latticectl",
			Short: "command line utility for interacting with lattices and systems",
			Subcommands: []latticectl.Command{
				&context.Command{
					Subcommands: []latticectl.Command{
						&context.GetCommand{},
						&context.SetCommand{},
					},
				},
				&bootstrap.Command{
					Subcommands: []latticectl.Command{
						&kubernetes.Command{},
					},
				},
				&local.Command{
					Subcommands: []latticectl.Command{
						&local.DownCommand{},
						&local.UpCommand{},
					},
				},
				&systems.ListSystemsCommand{
					Subcommands: []latticectl.Command{
						&systems.CreateCommand{},
						&systems.GetCommand{},
						&systems.DeleteCommand{},
						&systems.BuildCommand{},
						&systems.DeployCommand{},
						&systems.TeardownCommand{},
						&deploys.Command{
							Subcommands: []latticectl.Command{
								&deploys.GetCommand{},
							},
						},
						&secrets.Command{
							Subcommands: []latticectl.Command{
								&secrets.GetCommand{},
								&secrets.SetCommand{},
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
