package local

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

func Up() *cli.Command {
	var (
		id            string
		channel       string
		workDirectory string
		apiAuthKey    string
	)

	return &cli.Command{
		Flags: cli.Flags{
			"id": &flags.String{
				Default: "lattice",
				Target:  &id,
			},
			"container-channel": &flags.String{
				Default: "gcr.io/lattice-dev/laas/alpha",
				Target:  &channel,
			},
			"work-directory": &flags.String{
				Default: "/tmp/latticectl/local",
				Target:  &workDirectory,
			},
			"api-auth-key": &flags.String{
				Default: "",
				Target:  &apiAuthKey,
			},
		},
		Run: func(args []string, flags cli.Flags) error {
			return LocalUp(v1.LatticeID(id), channel, workDirectory, apiAuthKey, os.Stdout)
		},
	}
}

func LocalUp(id v1.LatticeID, containerChannel, workDirectory string, apiAuthKey string, w io.Writer) error {
	provisioner, err := local.NewLatticeProvisioner(workDirectory)
	if err != nil {
		return err
	}

	address, err := provisioner.Provision(id, containerChannel, apiAuthKey)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Lattice address:\n%v\n", address)
	return nil
}
