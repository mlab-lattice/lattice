package teardowns

import (
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListTeardownsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.TeardownCommand{
		Name: "status",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.TeardownCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Teardowns(ctx.SystemID())

			getFunc := func(client client.TeardownClient) ([]types.SystemTeardown, error) {
				teardown, err := client.Get(ctx.TeardownID())
				if err != nil {
					return nil, err
				}
				return []types.SystemTeardown{*teardown}, nil
			}

			if watch {
				WatchTeardowns(getFunc, c, format, os.Stdout)
				return
			}

			ListTeardowns(getFunc, c, format, os.Stdout)
		},
	}

	return cmd.Base()
}
