package teardowns

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

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
	output := &latticectl.OutputFlag{
		SupportedFormats: GetTeardownsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target:  &watch,
	}

	cmd := &latticectl.TeardownCommand{
		Name: "status",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.TeardownCommandContext, args []string) {
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

func GetTeardown(client v1client.TeardownClient, teardownID v1.TeardownID, format printer.Format, writer io.Writer) error {
	teardown, err := client.Get(teardownID)
	if err != nil {
		return err
	}

	p := teardownsPrinter([]v1.Teardown{*teardown}, format)
	p.Print(writer)
	return nil
}

func WatchTeardown(client v1client.TeardownClient, teardownID v1.TeardownID, format printer.Format, writer io.Writer) {
	teardowns := make(chan *v1.Teardown)

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
		p := teardownsPrinter([]v1.Teardown{*teardown}, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}
