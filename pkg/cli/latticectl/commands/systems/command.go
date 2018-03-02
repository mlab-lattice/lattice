package systems

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
	cmd := &latticectl.LatticeCommand{
		Name: "systems",
		Run: func(args []string, ctx latticectl.LatticeCommandContext) {
			c.run(ctx)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.BaseCommand()
}

func (c *Command) run(ctx latticectl.LatticeCommandContext) {
	systems, err := ctx.Client().Systems().List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", systems)
}

//type Command struct {
//	Subcommands []latticectl.Command
//	*latticectl.LatticeCommand
//}
//
//func (c *Command) Init() error {
//	c.LatticeCommand = &latticectl.LatticeCommand{
//		Name: "systems",
//		Run: func(args []string, ctx latticectl.LatticeCommandContext) {
//			c.run(ctx)
//		},
//		Subcommands: c.Subcommands,
//	}
//
//	return c.LatticeCommand.Init()
//}
//
//func (c *Command) run(ctx latticectl.LatticeCommandContext) {
//	systems, err := ctx.Client().Systems().List()
//	if err != nil {
//		log.Panic(err)
//	}
//
//	fmt.Printf("%v\n", systems)
//}
