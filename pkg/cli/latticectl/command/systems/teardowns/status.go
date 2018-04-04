package teardowns

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

// GetTeardownsSupportedFormats is the list of printer.Formats supported
// by the GetTeardown function.
var GetTeardownsSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: GetTeardownsSupportedFormats,
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

	p := teardownsPrinter([]types.SystemTeardown{*teardown}, format)
	p.Print(writer)
	return nil
}

func WatchTeardown(client client.TeardownClient, teardownID types.SystemTeardownID, format printer.Format, writer io.Writer) {
	teardowns := make(chan *types.SystemTeardown)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			teardown, err := client.Get(teardownID)
			if err != nil {
				return false, err
			}

			teardowns <- teardown
			return false, nil
		},
	)

	for teardown := range teardowns {
		p := teardownsPrinter([]types.SystemTeardown{*teardown}, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}
