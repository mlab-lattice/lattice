package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type BuildCommand struct {
	PreRun         func()
	ContextCreator func(ctx latticectl.LatticeCommandContext, systemID types.SystemID) latticectl.SystemCommandContext
	*latticectl.SystemCommand
}

func (c *BuildCommand) Init() error {
	var version string
	c.SystemCommand = &latticectl.SystemCommand{
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

	return c.SystemCommand.Init()
}

func (c *BuildCommand) run(ctx latticectl.SystemCommandContext, version string) {
	buildID, err := ctx.Client().Systems().SystemBuilds(ctx.SystemID()).Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
