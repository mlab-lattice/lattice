package local

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

type UpCommand struct {
}

func (c *UpCommand) Base() (*latticectl.BaseCommand, error) {
	var id string
	var channel string
	var workDirectory string
	var apiAuthKey string

	cmd := &latticectl.BaseCommand{
		Name: "up",
		Flags: cli.Flags{
			&flags.String{
				Name:    "id",
				Default: "lattice",
				Target:  &id,
			},
			&flags.String{
				Name:    "container-channel",
				Default: "gcr.io/lattice-dev/laas/alpha",
				Target:  &channel,
			},
			&flags.String{
				Name:    "work-directory",
				Default: "/tmp/latticectl/local",
				Target:  &workDirectory,
			},
			&flags.String{
				Name:    "api-auth-key",
				Default: "",
				Target:  &apiAuthKey,
			},
		},
		Run: func(lctl *latticectl.Latticectl, args []string) {
			Up(v1.LatticeID(id), channel, workDirectory, apiAuthKey)
		},
	}

	return cmd.Base()
}

func Up(id v1.LatticeID, containerChannel, workDirectory string, apiAuthKey string) {
	provisioner, err := local.NewLatticeProvisioner(workDirectory)
	if err != nil {
		log.Fatal(err)
	}

	address, err := provisioner.Provision(id, containerChannel, apiAuthKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Lattice address:\n%v\n", address)
}
