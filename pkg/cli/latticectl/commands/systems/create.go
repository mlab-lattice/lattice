package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type CreateCommand struct {
}

func (c *CreateCommand) BaseCommand() (*command.BaseCommand2, error) {
	var definitionURL string
	var systemName string
	cmd := &latticectl.LatticeCommand{
		Name: "create",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "definition",
				Required: true,
				Target:   &definitionURL,
			},
			&command.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &systemName,
			},
		},
		Run: func(args []string, ctx latticectl.LatticeCommandContext) {
			c.run(ctx, types.SystemID(systemName), definitionURL)
		},
	}

	return cmd.BaseCommand()
}

func (c *CreateCommand) run(ctx latticectl.LatticeCommandContext, name types.SystemID, definitionURL string) {
	system, err := ctx.Client().Systems().Create(name, definitionURL)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
