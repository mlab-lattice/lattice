package builds

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"sort"
	"bytes"
	"errors"

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

type GetCommand struct {
}

type PrintBuildState func(io.Writer, *spinner.Spinner, *types.SystemBuild)

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
				err = WatchBuild(c, ctx.BuildID(), format, os.Stdout, PrintBuildStateDuringWatchBuild)
				if err != nil {
					os.Exit(1)
				}
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
func WatchBuild(
	client client.SystemBuildClient,
	buildID types.SystemBuildID,
	format printer.Format,
	writer io.Writer,
	PrintBuildStateDuringWatchBuild PrintBuildState,
	) error {
	builds := make(chan *types.SystemBuild)
	
	lastHeight := 0
	var returnError error
	var exit bool
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
		p := BuildPrinter(build, format)
		lastHeight = p.Overwrite(b, lastHeight)
		
		if format == printer.FormatDefault || format == printer.FormatTable {
			PrintBuildStateDuringWatchBuild(writer, s, build)
		}
		
		exit, returnError = buildCompleted(build)
		if exit {
			return returnError
		}
	}
	
	return nil
}

func PrintBuildStateDuringWatchBuild(writer io.Writer, s *spinner.Spinner, build *types.SystemBuild)  {
	switch build.State {
	case types.SystemBuildStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(string(build.Version)))
	case types.SystemBuildStateRunning:
		s.Start()
		s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(string(build.Version)))
	case types.SystemBuildStateSucceeded:
		s.Stop()
		printBuildSuccess(writer, string(build.Version), build.ID)
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
		
		PrintBuildFailure(writer, string(build.Version), componentErrors)
	}
}

func buildCompleted(build *types.SystemBuild) (bool, error) {
	switch build.State {
	case types.SystemBuildStateSucceeded:
		return true, nil
	case types.SystemBuildStateFailed:
		return true, errors.New("System Build Failed")
	default:
		return false, nil
	}
}

func printBuildSuccess(writer io.Writer, version string, buildID types.SystemBuildID) {
	fmt.Fprint(writer, color.BoldHiSuccess("✓ "))
	fmt.Fprint(writer, color.BoldHiSuccess(version))
	fmt.Fprint(writer, color.BoldHiSuccess(" built successfully! You can now deploy this build using:\n"))
	fmt.Fprintf(writer, color.Success("\n    lattice systems:deploy %s\n"), buildID)
}

func PrintBuildFailure(writer io.Writer, version string, componentErrors [][]string) {
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
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
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
