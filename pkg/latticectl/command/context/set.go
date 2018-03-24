package context

import (
	"log"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/latticectl"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type SetCommand struct {
}

func (c *SetCommand) Base() (*latticectl.BaseCommand, error) {
	var lattice string
	var system string
	cmd := &latticectl.BaseCommand{
		Name: "set",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:     "lattice",
				Required: false,
				Target:   &lattice,
			},
			&cli.StringFlag{
				Name:     "system",
				Required: false,
				Target:   &system,
			},
		},
		Run: func(lctl *latticectl.Latticectl, args []string) {
			SetContext(lctl.Context, lattice, v1.SystemID(system))
		},
	}

	return cmd.Base()
}

func SetContext(ctxm latticectl.ContextManager, lattice string, system v1.SystemID) {
	if ctxm == nil {
		log.Fatal("cannot set context: no context manager set")
	}

	if lattice == "" && system != "" {
		ctx, err := ctxm.Get()
		if err != nil {
			log.Fatal(err)
		}

		lattice = ctx.Lattice()
		if lattice == "" {
			log.Fatal("cannot set --system without --lattice")
		}
	}

	if err := ctxm.Set(lattice, system); err != nil {
		log.Fatal(err)
	}
}
