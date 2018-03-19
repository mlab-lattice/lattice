package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type BuildCommand struct {
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	var version string
	cmd := &lctlcommand.SystemCommand{
		Name: "build",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			err := BuildSystem(ctx.Client().Systems().SystemBuilds(ctx.SystemID()), version)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func BuildSystem(client client.SystemBuildClient, version string) error {
	buildID, err := client.Create(version)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", buildID)
	return nil
}
