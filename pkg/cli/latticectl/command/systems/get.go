package systems

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"sort"
	"strings"

	"github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	
	tw "github.com/tfogo/tablewriter"
	//"github.com/briandowns/spinner"
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

// func WatchSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) {
// 	system, err := client.Get(systemID)
// 	if err != nil {
// 		log.Panic(err)
// 	}
//
// 	fmt.Printf("%v\n", system)
// }

func GetSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) error {
	system, err := client.Get(systemID)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", system)
	return nil
}

func WatchSystem(client client.SystemClient, systemID types.SystemID, format printer.Format, writer io.Writer) {
	// TODO: Currently the goroutines are synched up using channels.
	// We might be able to make the code cleaner with little/no performance impact
	// if we make the functions synchronous.
	
	printerChan := make(chan printer.Interface)
	printerRenderedChan := make(chan bool)
	systemChan := make(chan *types.System)
	
	// Poll the API for the builds and send it to the channel
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			system, err := client.Get(systemID)
			if err != nil {
				return false, err
			}

			p := SystemPrinter(system, format)
			printerChan <- p
			// NOTE: Blocks here until printerRenderedChan is read
			// then systemChan is available to be written to
			systemChan <- system
			return false, nil
		},
	)
	
	// If displaying a table, use the overwritting terminal watcher, if JSON
	// use the scrolling watcher
	var w printer.Watcher2
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		w = &printer.OverwrittingTerminalWatcher2{}
	case printer.FormatJSON:
		w = &printer.ScrollingWatcher2{}
	}

	w.Watch(printerChan, writer, printerRenderedChan)
	
	// Await termincating condition, also print accompanying text under the table
	// if format == printer.FormatDefault || format == printer.FormatTable {
	// 	printAccompanyingText2(writer, printerRenderedChan, systemChan)
	// } else {
	// 	stateWatcher2(writer, systemChan, printerRenderedChan)
	// }
}

// func stateWatcher2(writer io.Writer, systemChan chan *types.System, printerRenderedChan chan bool) {
// 	var sentSystem *types.System
//
// 	for {
// 		<-printerRenderedChan
// 		sentSystem = <-systemChan
// 		switch sentSystem.State {
// 		case types.SystemStateStable, types.SystemStateFailed:
// 			return
// 		}
// 	}
// }

// func printAccompanyingText2(writer io.Writer, printerRenderedChan chan bool, systemChan chan *types.System) {
// 	var sentSystem *types.System
// 	var s *spinner.Spinner
//
// 	for i := 0; true; i++ {
// 		<-printerRenderedChan
// 		sentSystem = <-systemChan
//
// 		if i == 0 {
// 			s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
// 			s.Start()
// 		}
//
// 		switch sentSystem.State {
// 		case types.SystemStateUpdating, types.SystemStateScaling, types.SystemStateDeleting:
// 			s.Suffix = fmt.Sprintf(" Build pending for version: ...")
// 		case types.SystemStateStable:
// 			s.Stop()
//
// 			//printBuildSuccess(writer, version, buildID)
// 			return
// 		case types.SystemStateFailed:
// 			s.Stop()
// 			/*
// 			var componentErrors [][]string
//
// 			for serviceName, service := range sentSystem.Services {
// 				for componentName, component := range service.Components {
// 					if component.State == types.ComponentBuildStateFailed {
// 						componentErrors = append(componentErrors, []string{
// 							fmt.Sprintf("%s:%s", serviceName, componentName),
// 							string(*component.FailureMessage),
// 						})
// 					}
// 				}
// 			}
// 			*/
//
// 			//printBuildFailure(writer, version, componentErrors)
// 			return
// 		}
// 	}
// }

func SystemPrinter(system *types.System, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Service", "State", "Info", "Updated", "Stale", "Addresses"}
		
		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}
		
		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
			{},
			{},
			{},
			{},
		}
		
		columnAlignment := []int{
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_LEFT,
			tw.ALIGN_RIGHT,
			tw.ALIGN_RIGHT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		// fmt.Fprintln(os.Stdout, system)
		for serviceName, service := range system.Services {
			// fmt.Fprintln(os.Stdout, service)
			
			// fmt.Fprintln(os.Stdout, component)
			// fmt.Fprint(os.Stdout, "COMPONENT STATE", component.State, "    ")
			var infoMessage string
			
			if service.FailureMessage == nil {
				infoMessage = ""
			} else {
				infoMessage = string(*service.FailureMessage)
			}
			
			var stateColor color.Color
			switch service.State {
			case types.ServiceStateStable:
				stateColor = color.Success
			case types.ServiceStateFailed:
				stateColor = color.Failure
			default:
				stateColor = color.Warning
			}
			
			var addresses []string
			for port, address := range service.PublicPorts {
				addresses = append(addresses, fmt.Sprintf("%v: %v", port, address))
			}
			
			rows = append(rows, []string{
				string(serviceName),
				stateColor(string(service.State)),
				string(infoMessage),
				string(service.UpdatedInstances),
				string(service.StaleInstances),
				strings.Join(addresses, ","),
			})
			
			sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
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
			Value: system,
		}
	}

	return p
}
