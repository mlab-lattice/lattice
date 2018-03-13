package services

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
)

// ListServicesSupportedFormats is the list of printer.Formats supported
// by the ListDeploys function.
var ListServicesSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type Command struct {
	Subcommands []latticectl.Command
}

func (c *Command) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.SystemCommand{
		Name: "deploys",
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

			if watch {
				WatchServices(ctx.Client().Systems().Services(ctx.SystemID()), format, os.Stdout)
			} else {
				ListServices(ctx.Client().Systems().Services(ctx.SystemID()), format, os.Stdout)
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListServices(client client.ServiceClient, format printer.Format, writer io.Writer) {
	deploys, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploys)
}

func WatchServices(client client.ServiceClient, format printer.Format, writer io.Writer) {
	//// Poll the API for the builds and send it to the channel
	//printerChan := make(chan printer.Interface)
	//go wait.PollImmediateInfinite(
	//	5*time.Second,
	//	func() (bool, error) {
	//		builds, err := client.List()
	//		if err != nil {
	//			return false, err
	//		}
	//
	//		p := buildsPrinter(builds, format)
	//		printerChan <- p
	//		return false, nil
	//	},
	//)
	//
	//// If displaying a table, use the overwritting terminal watcher, if JSON
	//// use the scrolling watcher
	//var w printer.Watcher
	//switch format {
	//case printer.FormatDefault, printer.FormatTable:
	//	w = &printer.OverwrittingTerminalWatcher{}
	//
	//case printer.FormatJSON:
	//	w = &printer.ScrollingWatcher{}
	//}
	//
	//w.Watch(printerChan, writer)
}
