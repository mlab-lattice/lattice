package local

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

func Down() *cli.Command {
	var (
		workDirectory string
	)

	return &cli.Command{
		Flags: cli.Flags{
			"work-directory": &flags.String{
				Default: "/tmp/latticectl/local",
				Target:  &workDirectory,
			},
		},
		Run: func(args []string, flags cli.Flags) error {
			return LocalDown(workDirectory)
		},
	}
}

func LocalDown(workDirectory string) error {
	provisioner, err := local.NewLatticeProvisioner(workDirectory)
	if err != nil {
		return err
	}

	if err := provisioner.Deprovision("default", true); err != nil {
		return err
	}

	return nil
}
