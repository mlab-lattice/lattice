package local

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type UpCommand struct {
}

func (c *UpCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var initialSystemDefinition string
	var registry string
	var channel string
	var workDirectory string

	cmd := &latticectl.BaseCommand{
		Name: "up",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:    "name",
				Default: "default",
				Target:  &name,
			},
			&cli.StringFlag{
				Name:   "initial-system-defintion",
				Target: &initialSystemDefinition,
			},
			&cli.StringFlag{
				Name:    "container-registry",
				Default: "gcr.io/lattice-dev",
				Target:  &registry,
			},
			&cli.StringFlag{
				Name:    "container-channel",
				Default: "stable-debug-",
				Target:  &channel,
			},
			&cli.StringFlag{
				Name:    "work-directory",
				Default: "/tmp/latticectl/local",
				Target:  &workDirectory,
			},
		},
		Run: func(lctl *latticectl.Latticectl, args []string) {
			Up(name, initialSystemDefinition, registry, channel, workDirectory)
		},
	}

	return cmd.Base()
}

func Up(name, initialSystemDefinition, registry, channel, workDirectory string) {
	provisioner, err := local.NewLatticeProvisioner(registry, channel, workDirectory, nil)
	if err != nil {
		log.Fatal(err)
	}

	var definition *string
	if initialSystemDefinition != "" {
		definition = &initialSystemDefinition
	}

	address, err := provisioner.Provision(name, definition)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Lattice address:\n%v\n", address)
}
