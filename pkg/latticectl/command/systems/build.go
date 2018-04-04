package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type BuildCommand struct {
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	var version string
	cmd := &command.SystemCommand{
		Name: "build",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(ctx command.SystemCommandContext, args []string) {
			BuildSystem(ctx.Client().Systems().Builds(ctx.SystemID()), v1.SystemVersion(version))
		},
	}

	return cmd.Base()
}

func BuildSystem(client clientv1.BuildClient, version v1.SystemVersion) {
	buildID, err := client.Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
