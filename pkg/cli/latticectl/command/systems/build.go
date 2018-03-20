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
			BuildSystem(ctx.Client().Systems().Builds(ctx.SystemID()), version)
		},
	}

	return cmd.Base()
}

func BuildSystem(client client.BuildClient, version string) {
	buildID, err := client.Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
