package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/latticectl/command"
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
			BuildSystem(ctx.Client().Systems().Builds(ctx.SystemID()), version)
		},
	}

	return cmd.Base()
}

func BuildSystem(client clientv1.BuildClient, version string) {
	buildID, err := client.Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
