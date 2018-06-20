package builds

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/briandowns/spinner"
	tw "github.com/tfogo/tablewriter"
)

type StatusCommand struct {
}

type PrintBuildState func(io.Writer, *spinner.Spinner, *v1.Build)

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListBuildsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.BuildCommand{
		Name: "status",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.BuildCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Builds(ctx.SystemID())

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

func GetBuild(client v1client.BuildClient, buildID v1.BuildID, format printer.Format, writer io.Writer) error {
	build, err := client.Get(buildID)
	if err != nil {
		return err
	}

	p := BuildPrinter(build, format)
	p.Print(writer)
	return nil
}

// func WatchBuild(v1client v1client.SystemBuildClient, buildID v1.SystemBuildID, format printer.Format, writer io.Writer) {
// 	build, err := v1client.Get(buildID)
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
	client v1client.BuildClient,
	buildID v1.BuildID,
	format printer.Format,
	writer io.Writer,
	PrintBuildStateDuringWatchBuild PrintBuildState,
) error {
	builds := make(chan *v1.Build)

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

		if format == printer.FormatTable {
			PrintBuildStateDuringWatchBuild(writer, s, build)
		}

		exit, returnError = buildCompleted(build)
		if exit {
			return returnError
		}
	}

	return nil
}

func PrintBuildStateDuringWatchBuild(writer io.Writer, s *spinner.Spinner, build *v1.Build) {
	switch build.State {
	case v1.BuildStatePending:
		s.Start()
		s.Suffix = fmt.Sprintf(" Build pending for version: %s...", color.ID(string(build.Version)))
	case v1.BuildStateRunning:
		s.Start()
		s.Suffix = fmt.Sprintf(" Building version: %s...", color.ID(string(build.Version)))
	case v1.BuildStateSucceeded:
		s.Stop()
		printBuildSuccess(writer, string(build.Version), build.ID)
	case v1.BuildStateFailed:
		s.Stop()

		var componentErrors [][]string

		for serviceName, service := range build.Services {
			for componentName, component := range service.Components {
				if component.State == v1.ComponentBuildStateFailed {
					componentErrorMessage := ""

					if component.FailureMessage != nil {
						componentErrorMessage = *component.FailureMessage
					}

					componentErrors = append(componentErrors, []string{
						fmt.Sprintf("%s:%s", serviceName, componentName),
						componentErrorMessage,
					})
				}
			}
		}

		PrintBuildFailure(writer, string(build.Version), componentErrors)
	}
}

func buildCompleted(build *v1.Build) (bool, error) {
	switch build.State {
	case v1.BuildStateSucceeded:
		return true, nil
	case v1.BuildStateFailed:
		return true, errors.New("System Build Failed")
	default:
		return false, nil
	}
}

func printBuildSuccess(writer io.Writer, version string, buildID v1.BuildID) {
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

func BuildPrinter(build *v1.Build, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		headers := []string{"Component", "State", "Started At", "Completed At", "Info"}

		headerColors := []tw.Colors{
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
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
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
				case v1.ComponentBuildStateSucceeded:
					stateColor = color.Success
				case v1.ComponentBuildStateFailed:
					stateColor = color.Failure
				default:
					stateColor = color.Warning
				}

				startTimestamp := ""
				completionTimestamp := ""

				if component.StartTimestamp != nil {
					startTimestamp = component.StartTimestamp.String()
				}

				if component.CompletionTimestamp != nil {
					completionTimestamp = component.CompletionTimestamp.String()
				}

				rows = append(rows, []string{
					fmt.Sprintf("%s:%s", serviceName, componentName),
					stateColor(string(component.State)),
					startTimestamp,
					completionTimestamp,
					string(infoMessage),
				})

				sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
			}
		}

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: build,
		}
	}

	return p
}
