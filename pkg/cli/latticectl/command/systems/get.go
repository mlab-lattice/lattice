package systems

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
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
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
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems()

			if watch {
				WatchSystem(c, ctx.SystemID(), format, os.Stdout)
				return
			}

			GetSystem(c, ctx.SystemID(), format, os.Stdout)
		},
	}

	return cmd.Base()
}

func WatchSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) {
	system, err := client.Get(systemID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}

func GetSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) error {
	system, err := client.Get(systemID)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", system)
	return nil
}
