package command

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

type LatticeCommandContext struct {
	Context *Context
	Lattice string
	Client  client.Interface
}

type LatticeCommand struct {
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx *LatticeCommandContext, args []string, flags cli.Flags)
	Subcommands map[string]*cli.Command
}

func (c *LatticeCommand) Command() *cli.Command {
	c.Flags["lattice"] = &flags.String{
		Required: false,
	}

	cmd := &cli.Command{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		Run: func(args []string) {
			configFile := &configFile{}
			context, err := configFile.GetContext()
			if err != nil {
				// TODO(kevindrosendahl): better handling here
				panic(err)
			}

			lattice := c.Flags["lattice"].Value().(string)
			// Try to retrieve the lattice from the context if there is one
			if lattice == "" {
				lattice = context.Lattice
			}

			if lattice == "" {
				log.Fatal("required flag lattice must be set")
			}

			ctx := &LatticeCommandContext{
				Context: context,
				Lattice: lattice,
				// FIXME(kevindrosendahl): support api auth key
				Client: rest.NewClient(lattice, ""),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd
}
