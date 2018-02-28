package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeployCommand struct {
	PreRun func()
	*latticectl.SystemCommand
}

func (c *DeployCommand) Init() error {
	var buildID string
	var version string
	c.SystemCommand = &latticectl.SystemCommand{
		Name: "deploy",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "build",
				Required: true,
				Target:   &buildID,
			},
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx, types.SystemBuildID(buildID), version)
		},
	}

	return c.SystemCommand.Init()
}

func (c *DeployCommand) run(
	ctx latticectl.SystemCommandContext,
	buildID types.SystemBuildID,
	version string,
) {
	if buildID == "" && version == "" {
		log.Panic("must provide either build or version")
	}

	systemID := ctx.SystemID()

	var deployID types.SystemRolloutID
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deployID, err = ctx.Client().Systems().Rollouts(systemID).CreateFromBuild(buildID)
		}
	} else {
		deployID, err = ctx.Client().Systems().Rollouts(systemID).CreateFromVersion(version)
	}

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deployID)
}
