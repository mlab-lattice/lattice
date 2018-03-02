package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type BuildCommand struct {
}

func (c *BuildCommand) BaseCommand() (*command.BaseCommand2, error) {
	var version string
	cmd := &latticectl.SystemCommand{
		Name: "build",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx, version)
		},
	}

	return cmd.BaseCommand()
}

func (c *BuildCommand) run(ctx latticectl.SystemCommandContext, version string) {
	buildID, err := ctx.Client().Systems().SystemBuilds(ctx.SystemID()).Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
