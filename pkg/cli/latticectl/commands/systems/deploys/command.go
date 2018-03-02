package deploys

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
)

type Command struct {
	Subcommands []command.Command2
}

func (c *Command) BaseCommand() (*command.BaseCommand2, error) {
	cmd := &latticectl.SystemCommand{
		Name: "deploys",
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.BaseCommand()
}

func (c *Command) run(ctx latticectl.SystemCommandContext) {
	deploys, err := ctx.Client().Systems().Rollouts(ctx.SystemID()).List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploys)
}
