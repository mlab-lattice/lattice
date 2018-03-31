package builds

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"sort"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	
	tw "github.com/tfogo/tablewriter"
	"github.com/briandowns/spinner"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListBuildsSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.BuildCommand{
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
		Run: func(ctx lctlcommand.BuildCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().SystemBuilds(ctx.SystemID())

			if watch {
				WatchBuild(c, ctx.BuildID(), format, os.Stdout)
			}

			err = GetBuild(c, ctx.BuildID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetBuild(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer) error {
	build, err := client.Get(buildID)
	if err != nil {
		return err
	}

	p := BuildPrinter(build, format)
	p.Print(writer)
	return nil
}

// func WatchBuild(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer) {
// 	build, err := client.Get(buildID)
// 	if err != nil {
// 		log.Panic(err)
// 	}
//
// 	fmt.Printf("%v\n", build)
// }

// WatchBuilds polls the API for the current Builds, and writes out the Builds to the
// the supplied io.Writer in the given printer.Format, unless the printer.Format is
// printer.FormatTable, in which case it always writes to the terminal.
func WatchBuild(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer) {
	// Poll the API for the builds and send it to the channel
	printerChan := make(chan printer.Interface)
	spinnerChan := make(chan bool)
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			build, err := client.Get(buildID)
			if err != nil {
				return false, err
			}

			p := BuildPrinter(build, format)
			printerChan <- p
			return false, nil
		},
	)

	// If displaying a table, use the overwritting terminal watcher, if JSON
	// use the scrolling watcher
	var w printer.Watcher2
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		w = &printer.OverwrittingTerminalWatcher2{}

	// case printer.FormatJSON:
	// 	w = &printer.ScrollingWatcher{}
	}
	
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Building..."

	go w.Watch(printerChan, writer, spinnerChan)
	
	if <-spinnerChan {
		s.Start()
	}
	
	// Fix up
	<-spinnerChan
	
}

func BuildPrinter(build *types.SystemBuild, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Component", "State", "Info"}
		
		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}
		
		columnColors := []tw.Colors{
			{tw.FgHiMagentaColor},
			{},
			{},
		}
		
		columnAlignment := []int{
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for serviceName, service := range build.Services {
			for componentName, component := range service.Components {
				
				var infoMessage string
				
				if component.FailureMessage == nil {
					infoMessage = string(*component.LastObservedPhase)
				} else {
					infoMessage = string(*component.FailureMessage)
				}
				
				rows = append(rows, []string{
					fmt.Sprintf("%s:%s", serviceName, componentName),
					string(component.State),
					string(infoMessage),
				})
				
				sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
			}
		}
		
		p = &printer.Table{
			Headers: 					headers,
			Rows:    					rows,
			HeaderColors: 		headerColors,
			ColumnColors: 		columnColors,
			ColumnAlignment: 	columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: build,
		}
	}

	return p
}
