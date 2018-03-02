package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) BaseCommand() (*command.BaseCommand2, error) {
	var definitionURL string
	var systemName string
	cmd := &latticectl.SystemCommand{
		Name: "get",
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx, types.SystemID(systemName), definitionURL)
		},
	}

	return cmd.BaseCommand()
}

func (c *GetCommand) run(ctx latticectl.SystemCommandContext, name types.SystemID, definitionURL string) {
	system, err := ctx.Client().Systems().Get(ctx.SystemID())
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
