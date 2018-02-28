package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
	PreRun func()
	*latticectl.SystemCommand
}

func (c *GetCommand) Init() error {
	var definitionURL string
	var systemName string
	c.SystemCommand = &latticectl.SystemCommand{
		Name: "get",
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx, types.SystemID(systemName), definitionURL)
		},
	}

	return c.SystemCommand.Init()
}

func (c *GetCommand) run(ctx latticectl.SystemCommandContext, name types.SystemID, definitionURL string) {
	system, err := ctx.Client().Systems().Get(ctx.SystemID())
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
