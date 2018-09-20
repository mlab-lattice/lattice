package local

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type DownCommand struct {
}

func (c *DownCommand) Base() (*latticectl.BaseCommand, error) {
	var name string
	var workDirectory string

	cmd := &latticectl.BaseCommand{
		Name: "down",
		Flags: cli.Flags{
			&flags.String{
				Name:    "name",
				Default: "default",
				Target:  &name,
			},
			&flags.String{
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
	provisioner, err := local.NewLatticeProvisioner(workDirectory)
	if err != nil {
		log.Fatal(err)
	}

	if err := provisioner.Deprovision(name, true); err != nil {
		log.Fatal(err)
	}
}
