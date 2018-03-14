package local

import (
	"log"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type DownCommand struct {
}

func (c *DownCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var workDirectory string

	cmd := &latticectl.BaseCommand{
		Name: "down",
		Flags: command.Flags{
			&command.StringFlag{
				Name:    "name",
				Default: "default",
				Target:  &name,
			},
			&command.StringFlag{
				Name:    "work-directory",
				Default: "/tmp/latticectl/local",
				Target:  &workDirectory,
			},
		},
		Run: func(lctl *latticectl.Latticectl, args []string) {
			Down(name, workDirectory)
		},
	}

	return cmd.Base()
}

func Down(name, workDirectory string) {
	provisioner, err := local.NewClusterProvisioner("", "", workDirectory, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := provisioner.Deprovision(name, true); err != nil {
		log.Fatal(err)
	}
}
