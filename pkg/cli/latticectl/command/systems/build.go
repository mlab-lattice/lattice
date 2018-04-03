package systems

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"sort"
	"bytes"
	"strings"

	"github.com/buger/goterm"

	"github.com/mlab-lattice/system/pkg/cli/color"
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

type BuildCommand struct {}

type PrintBuildState func(io.Writer, *spinner.Spinner, *types.SystemBuild, string)

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListSystemsSupportedFormats,
	}
	var watch bool
	var version string
	
	cmd := &lctlcommand.SystemCommand{
		Name: "build",
		Flags: []command.Flag{
			output.Flag(),
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
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
			
			c := ctx.Client().Systems().SystemBuilds(ctx.SystemID())
			
			err = BuildSystem(c, version, format, os.Stdout, watch)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func BuildSystem(
	client client.SystemBuildClient,
	version string,
	format printer.Format,
	writer io.Writer,
	watch bool,
	) error {
	buildID, err := client.Create(version)
	if err != nil {
		return err
	}
	
	if watch {
		if format == printer.FormatDefault || format == printer.FormatTable {
			fmt.Fprintf(writer, "\nBuild ID: %s\n", color.ID(string(buildID)))
		}
		WatchBuild(client, buildID, format, os.Stdout, version, printBuildState)
	} else {
		fmt.Fprintf(writer, "Building version %s, Build ID: %s\n\n", version, color.ID(string(buildID)))
		fmt.Fprintf(writer, "To view the status of the build, run:\n\n    latticectl system:builds:status --build %s [--watch]\n", color.ID(string(buildID)))
	}
	return nil
}

func WatchBuild(
	client client.SystemBuildClient,
	buildID types.SystemBuildID,
	format printer.Format,
	writer io.Writer,
	version string,
	printBuildState PrintBuildState,
	) {
	builds := make(chan *types.SystemBuild)
	
	lastHeight := 0
	var b bytes.Buffer
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	
	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			build, err := client.Get(buildID)
			if err != nil {
				return false, err
			}

			builds <- build
			return false, nil
		},
	)
	
	for build := range builds {
		lastHeight = printOutput(b, lastHeight, build, format)
		
		if format == printer.FormatDefault || format == printer.FormatTable {
			printBuildState(writer, s, build, version)
		}
		
		if buildCompleted(build) {
			return
		}
	}
}

func printBuildState(writer io.Writer, s *spinner.Spinner, build *types.SystemBuild, version string)  {
	switch build.State {
	case types.SystemBuildStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(version))
	case types.SystemBuildStateRunning:
		s.Start()
		s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(version))
	case types.SystemBuildStateSucceeded:
		s.Stop()
		printBuildSuccess(writer, version, build.ID)
	case types.SystemBuildStateFailed:
		s.Stop()
		
		var componentErrors [][]string
		
		for serviceName, service := range build.Services {
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
	}
}

func buildCompleted(build *types.SystemBuild) bool {
	switch build.State {
	case types.SystemBuildStateSucceeded, types.SystemBuildStateFailed:
		return true
	default:
		return false
	}
}

func printOutput(
	b bytes.Buffer,
	lastHeight int,
	build *types.SystemBuild,
	format printer.Format,
	) int {
	printer := BuildPrinter(build, format)

	// Read the new printer's output
	printer.Print(&b)
	output := b.String()

	// Remove the new printer's output from the buffer
	b.Truncate(0)

	for i := 0; i <= lastHeight; i++ {
		if i != 0 {
			goterm.MoveCursorUp(1)
			goterm.ResetLine("")
		}
	}

	goterm.Print(output)
	goterm.Flush()

	return len(strings.Split(output, "\n"))
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
			{tw.FgHiCyanColor},
			{},
			{},
		}
		
		columnAlignment := []int{
			tw.ALIGN_CENTER,
			tw.ALIGN_CENTER,
			tw.ALIGN_LEFT,
		}

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
