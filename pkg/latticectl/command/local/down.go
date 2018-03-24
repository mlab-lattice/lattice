package local

import (
	"log"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type DownCommand struct {
}

func (c *DownCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var workDirectory string

	cmd := &latticectl.BaseCommand{
		Name: "down",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:    "name",
				Default: "default",
				Target:  &name,
			},
			&cli.StringFlag{
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
	provisioner, err := local.NewLatticeProvisioner("", "", workDirectory, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := provisioner.Deprovision(name, true); err != nil {
		log.Fatal(err)
	}
}
