package builds

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

func Status() *cli.Command {
	var (
		output string
		watch  bool
	)

	cmd := Command{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		Run: func(ctx *BuildCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)

			if watch {
				WatchBuild(ctx.Client, ctx.System, ctx.Build, os.Stdout, format)
				return nil
			}

			return PrintBuild(ctx.Client, ctx.System, ctx.Build, os.Stdout, format)
		},
	}

	return cmd.Command()
}

func PrintBuild(client client.Interface, system v1.SystemID, id v1.BuildID, w io.Writer, f printer.Format) error {
	build, err := client.V1().Systems().Builds(system).Get(id)
	if err != nil {
		return err
	}

	switch f {
	case printer.FormatTable:
		dw := buildWriter(w)
		s := buildString(build)
		dw.Print(s)

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		j.Print(build)

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	return nil
}

func WatchBuild(client client.Interface, system v1.SystemID, id v1.BuildID, w io.Writer, f printer.Format) error {
	var handle func(*v1.Build) bool
	switch f {
	case printer.FormatTable:
		dw := buildWriter(w)

		handle = func(build *v1.Build) bool {
			s := buildString(build)
			dw.Overwrite(s)

			switch build.Status.State {
			case v1.BuildStateFailed:
				fmt.Fprint(w, color.BoldHiSuccessString("✘ build failed\n"))
				return true

			case v1.BuildStateSucceeded:
				fmt.Fprint(w, color.BoldHiSuccessString("✓ build succeeded\n"))
				return true
			}

			return false
		}

	case printer.FormatJSON:
		j := printer.NewJSON(w)
		handle = func(build *v1.Build) bool {
			j.Print(build)
			return false
		}

	default:
		return fmt.Errorf("unexpected format %v", f)
	}

	for {
		build, err := client.V1().Systems().Builds(system).Get(id)
		if err != nil {
			return err
		}

		done := handle(build)
		if done {
			return nil
		}

		time.Sleep(5 * time.Nanosecond)
	}
}

func buildWriter(w io.Writer) *printer.Custom {
	return printer.NewCustom(w)
}

func buildString(build *v1.Build) string {
	var spec string
	switch {
	case build.Path != nil:
		spec = fmt.Sprintf("path %s", build.Path.String())

	case build.Version != nil:
		spec = fmt.Sprintf("version %s", *build.Version)
	}

	stateColor := color.BoldString
	switch build.Status.State {
	case v1.BuildStatePending, v1.BuildStateAccepted, v1.BuildStateRunning:
		stateColor = color.BoldHiWarningString

	case v1.BuildStateSucceeded:
		stateColor = color.BoldHiSuccessString

	case v1.BuildStateFailed:
		stateColor = color.BoldHiFailureString
	}

	additional := ""
	if build.Status.Message != "" {
		additional += fmt.Sprintf(`
  message: %s`,
			build.Status.Message,
		)
	}

	if build.Status.StartTimestamp != nil {
		additional += fmt.Sprintf(`
  started: %s`,
			build.Status.StartTimestamp.String(),
		)
	}

	if build.Status.CompletionTimestamp != nil {
		additional += fmt.Sprintf(`
  completed: %s`,
			build.Status.CompletionTimestamp.String(),
		)
	}

	if build.Status.Path != nil {
		additional += fmt.Sprintf(`
  path: %s`,
			build.Status.Path.String(),
		)
	}

	if build.Status.Version != nil {
		additional += fmt.Sprintf(`
  version: %s`,
			string(*build.Status.Version),
		)
	}

	if len(build.Status.Workloads) != 0 {
		additional += `
  workloads:`
	}
	var paths []string
	for path := range build.Status.Workloads {
		paths = append(paths, path.String())
	}

	sort.Strings(paths)
	for _, p := range paths {
		path, _ := tree.NewPath(p)
		workload := build.Status.Workloads[path]

		mainDescriptor := ""
		if len(workload.Sidecars) != 0 {
			mainDescriptor = " (main container)"
		}
		mainColor := containerBuildColor(workload.Status.State)
		additional += mainColor(
			fmt.Sprintf(`
    %v%v`,
				path,
				mainDescriptor,
			),
		)

		for sidecar, sidecarBuild := range workload.Sidecars {
			sidecarColor := containerBuildColor(sidecarBuild.Status.State)
			additional += sidecarColor(
				fmt.Sprintf(`
    %v (%v sidecar)`,
					path,
					sidecar,
				),
			)
		}
	}

	return fmt.Sprintf(`build %v (%v)
  state: %v%v
`,
		color.IDString(string(build.ID)),
		spec,
		stateColor(string(build.Status.State)),
		additional,
	)
}

func containerBuildColor(state v1.ContainerBuildState) color.Formatter {
	switch state {
	case v1.ContainerBuildStateSucceeded:
		return color.SuccessString

	case v1.ContainerBuildStateFailed:
		return color.FailureString

	default:
		return color.WarningString
	}
}
