package latticectl

import (
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
)

type LatticeCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx LatticeCommandContext)
	Subcommands []command.Command2
}

func (c *LatticeCommand) BaseCommand() (*command.BaseCommand2, error) {
	var lattice string
	latticeURLFlag := &command.StringFlag{
		Name:     "lattice",
		Required: false,
		Target:   &lattice,
	}
	flags := append(c.Flags, latticeURLFlag)

	cmd := &BaseCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string, ctxm ContextManager, clientGenerator LatticeClientGenerator) {
			// Try to retrieve the lattice from the context if there is one
			if lattice == "" && ctxm != nil {
				ctx, err := ctxm.Get()
				if err != nil {
					panic(err)
				}

				lattice = ctx.Lattice()
			}

			if clientGenerator == nil {
				log.Fatal("client generator must be set")
			}

			if lattice == "" {
				log.Fatal("required flag lattice must be set")
			}

			ctx := &latticeCommandContext{
				lattice:       lattice,
				latticeClient: clientGenerator(lattice),
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.BaseCommand()
}

//
//type LatticeCommand struct {
//	Name        string
//	Short       string
//	Args        command.Args
//	Flags       command.Flags
//	PreRun      func()
//	Run         func(args []string, ctx LatticeCommandContext)
//	Subcommands []command.Command2
//	*BaseCommand
//}
//
//func (c *LatticeCommand) Init() error {
//	var lattice string
//	latticeURLFlag := &command.StringFlag{
//		Name:     "lattice",
//		Required: false,
//		Target:   &lattice,
//	}
//	flags := append(c.Flags, latticeURLFlag)
//
//	var subcommands []Command
//	for _, subcommand := range c.Subcommands {
//		subcommands = append(subcommands, subcommand)
//	}
//
//	c.BaseCommand = &BaseCommand{
//		Name:   c.Name,
//		Short:  c.Short,
//		Args:   c.Args,
//		Flags:  flags,
//		PreRun: c.PreRun,
//		Run: func(args []string) {
//			// Try to retrieve the lattice from the context if there is one
//			if lattice == "" && c.Context != nil {
//				ctx, err := c.Context.Get()
//				if err != nil {
//					panic(err)
//				}
//
//				lattice = ctx.Lattice()
//			}
//
//			if lattice == "" {
//				log.Fatal("required flag lattice must be set")
//			}
//
//			ctx := &latticeCommandContext{
//				lattice:       lattice,
//				latticeClient: c.Client(lattice),
//			}
//			c.Run(args, ctx)
//		},
//		Subcommands: subcommands,
//	}
//
//	return c.BaseCommand.Init()
//}
