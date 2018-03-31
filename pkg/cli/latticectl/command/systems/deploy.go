package systems

import (
	"fmt"
	"io"
	"log"
	"os"
	// "time"
	// "sort"

	// "github.com/mlab-lattice/system/pkg/cli/color"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	// "k8s.io/apimachinery/pkg/util/wait"
	//"github.com/mlab-lattice/system/pkg/cli/latticectl/command/systems"
	
	// tw "github.com/tfogo/tablewriter"
	// "github.com/briandowns/spinner"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	var buildID string
	var version string
	cmd := &lctlcommand.SystemCommand{
		Name: "deploy",
		Flags: []command.Flag{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
			&command.StringFlag{
				Name:     "build",
				Required: false,
				Target:   &buildID,
			},
			&command.StringFlag{
				Name:     "version",
				Required: false,
				Target:   &version,
			},
		},
		Run: func(ctx lctlcommand.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}
			
			err = DeploySystem(ctx.Client().Systems(), ctx.SystemID(), types.SystemBuildID(buildID), version, os.Stdout, format)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client client.SystemClient,
	systemID types.SystemID,
	buildID types.SystemBuildID,
	version string,
	writer io.Writer,
	format printer.Format,
) error {
	if buildID == "" && version == "" {
		return fmt.Errorf("must provide either build or version")
	}

	var deployID types.SystemRolloutID
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deployID, err = client.Rollouts(systemID).CreateFromBuild(buildID)
		}
	} else {
		deployID, err = client.Rollouts(systemID).CreateFromVersion(version)
	}

	if err != nil {
		return err
	}
	
	//TODO: Could reduce the number of requests necessary by
	// changing the behaviour of the client to return the
	// whole deploy on creation.
	deploy, err := client.Rollouts(systemID).Get(deployID)
	if err != nil {
		return err
	}
	
	build, err := client.SystemBuilds(systemID).Get(deploy.BuildID)
	if err != nil {
		return err
	}
	
 	//WatchBuild(client.SystemBuilds(systemID), deploy.BuildID, "table", writer, version)
	WatchSystem(client, systemID, format, os.Stdout)
	
	//p.Print(writer)

	fmt.Printf("%v\n", deployID, deploy, build)
	return nil
}


/*
func WatchDeploy(client client.SystemBuildClient, buildID types.SystemBuildID, format printer.Format, writer io.Writer, version string) {
	// TODO: Currently the goroutines are synched up using channels.
	// We might be able to make the code cleaner with little/no performance impact
	// if we make the functions synchronous.
	
	printerChan := make(chan printer.Interface)
	printerRenderedChan := make(chan bool)
	buildChan := make(chan *types.SystemBuild)
	
	// Poll the API for the builds and send it to the channel
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			build, err := client.Get(buildID)
			if err != nil {
				return false, err
			}

			p := buildPrinter(build, format)
			printerChan <- p
			// NOTE: Blocks here until printerRenderedChan is read
			// then buildChan is available to be written to
			buildChan <- build
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
	if format == printer.FormatDefault || format == printer.FormatTable {
		printAccompanyingText(writer, printerRenderedChan, buildChan, version, buildID)
	} else {
		stateWatcher(writer, buildChan, printerRenderedChan)
	}
}

/*
func stateWatcher(writer io.Writer, buildChan chan *types.SystemBuild, printerRenderedChan chan bool) {
	var sentBuild *types.SystemBuild
	
	for {
		<-printerRenderedChan
		sentBuild = <-buildChan
		switch sentBuild.State {
		case types.SystemBuildStateSucceeded, types.SystemBuildStateFailed:
			return
		}
	}
}

func printAccompanyingText(writer io.Writer, printerRenderedChan chan bool, buildChan chan *types.SystemBuild, version string, buildID types.SystemBuildID) {
	var sentBuild *types.SystemBuild
	var s *spinner.Spinner

	for i := 0; true; i++ {
		<-printerRenderedChan
		sentBuild = <-buildChan
		
		if i == 0 {
			s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Start()
		}
		
		switch sentBuild.State {
		case types.SystemBuildStatePending:
			s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(version))
		case types.SystemBuildStateRunning:
			s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(version))
		case types.SystemBuildStateSucceeded:
			s.Stop()
			
			printBuildSuccess(writer, version, buildID)
			return
		case types.SystemBuildStateFailed:
			s.Stop()
			
			var componentErrors [][]string
			
			for serviceName, service := range sentBuild.Services {
				for componentName, component := range service.Components {
					if component.State == types.ComponentBuildStateFailed {
						componentErrors = append(componentErrors, []string{
							fmt.Sprintf("%s:%s", serviceName, componentName),
							string(*component.FailureMessage),
						})
					}
				}
			}
			
			printBuildFailure(writer, version, componentErrors)
			return
		}
	}
}

func printBuildSuccess(writer io.Writer, version string, buildID types.SystemBuildID) {
	fmt.Fprint(writer, color.BoldHiSuccess("✓ "))
	fmt.Fprint(writer, color.BoldHiSuccess(version))
	fmt.Fprint(writer, color.BoldHiSuccess(" built successfully! You can now deploy this build using:\n"))
	fmt.Fprintf(writer, color.Success("\n    lattice systems:deploy %s\n"), buildID)
}

func printBuildFailure(writer io.Writer, version string, componentErrors [][]string) {
	fmt.Fprint(writer, color.BoldHiFailure("✘ Error building version "))
	fmt.Fprint(writer, color.BoldHiFailure(version))
	fmt.Fprint(writer, color.BoldHiFailure(":\n\n"))
	for _, componentError := range componentErrors {
		fmt.Fprintf(writer, color.Failure("Error building component %s, Error message:\n\n    %s\n"), componentError[0], componentError[1])
	}
}
*/

/*
func deployPrinter(build *types.SystemBuild, format printer.Format) printer.Interface {
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
			{tw.FgHiCyanColor},
			{},
			{},
		}
		
		columnAlignment := []int{
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_LEFT,
		}
		fmt.Fprintln(os.Stderr, "buildPrinter BUILD: ", build, "\n\n-------\n\n")

		var rows [][]string
		// fmt.Fprintln(os.Stdout, build)
		for serviceName, service := range build.Services {
			// fmt.Fprintln(os.Stdout, service)
			for componentName, component := range service.Components {
				// fmt.Fprintln(os.Stdout, component)
				//fmt.Fprint(os.Stdout, "COMPONENT STATE", component.State, "    ")
				var infoMessage string
				
				if component.FailureMessage == nil {
					if component.LastObservedPhase != nil {
						infoMessage = string(*component.LastObservedPhase)
					} else {
						infoMessage = ""
					}
				} else {
					infoMessage = string(*component.FailureMessage)
				}
				
				var stateColor color.Color
				switch component.State {
				case types.ComponentBuildStateSucceeded:
					stateColor = color.Success
				case types.ComponentBuildStateFailed:
					stateColor = color.Failure
				default:
					stateColor = color.Warning
				}
				
				rows = append(rows, []string{
					fmt.Sprintf("%s:%s", serviceName, componentName),
					stateColor(string(component.State)),
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
*/
