package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type GetCommand struct {
	PreRun func()
	*latticectl.DeployCommand
}

func (c *GetCommand) Init() error {
	c.DeployCommand = &latticectl.DeployCommand{
		Name:   "get",
		PreRun: c.PreRun,
		Run: func(args []string, ctx latticectl.DeployCommandContext) {
			c.run(ctx)
		},
	}

	return c.DeployCommand.Init()
}

func (c *GetCommand) run(ctx latticectl.DeployCommandContext) {
	deploy, err := ctx.Client().Systems().Rollouts(ctx.SystemID()).Get(ctx.DeployID())
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}
