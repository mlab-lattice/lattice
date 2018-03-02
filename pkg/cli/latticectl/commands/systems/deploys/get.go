package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type GetCommand struct {
}

func (c *GetCommand) BaseCommand() (*command.BaseCommand2, error) {
	cmd := &latticectl.DeployCommand{
		Name: "get",
		Run: func(args []string, ctx latticectl.DeployCommandContext) {
			c.run(ctx)
		},
	}

	return cmd.BaseCommand()
}

func (c *GetCommand) run(ctx latticectl.DeployCommandContext) {
	deploy, err := ctx.Client().Systems().Rollouts(ctx.SystemID()).Get(ctx.DeployID())
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}
