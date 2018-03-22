package teardowns

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
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

			if watch {
				WatchTeardown(c, ctx.TeardownID(), format, os.Stdout)
				return
			}

			err = GetTeardown(c, ctx.TeardownID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetTeardown(client client.TeardownClient, teardownID types.SystemTeardownID, format printer.Format, writer io.Writer) error {
	teardown, err := client.Get(teardownID)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", teardown)
	return nil
}

func WatchTeardown(client client.TeardownClient, teardownID types.SystemTeardownID, format printer.Format, writer io.Writer) {
	teardown, err := client.Get(teardownID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", teardown)
}
