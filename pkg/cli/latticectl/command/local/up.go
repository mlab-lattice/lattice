package local

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
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
		Flags: command.Flags{
			&command.StringFlag{
				Name:    "name",
				Default: "default",
				Target:  &name,
			},
			&command.StringFlag{
				Name:   "initial-system-defintion",
				Target: &initialSystemDefinition,
			},
			&command.StringFlag{
				Name:    "container-registry",
				Default: "gcr.io/lattice-dev",
				Target:  &registry,
			},
			&command.StringFlag{
				Name:    "container-channel",
				Default: "stable-debug-",
				Target:  &channel,
			},
			&command.StringFlag{
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
	provisioner, err := local.NewClusterProvisioner(registry, channel, workDirectory, nil)
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
