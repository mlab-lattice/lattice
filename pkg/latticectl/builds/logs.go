package builds

import (
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

// Logs returns a Command to retrieve the logs of a build.
func Logs() *cli.Command {
	var (
		follow     bool
		path       tree.Path
		previous   bool
		sidecar    string
		timestamps bool
		since      string
		tail       int
	)

	cmd := Command{
		Flags: map[string]cli.Flag{
			"follow": &flags.Bool{Target: &follow},
			"path": &flags.Path{
				Required: true,
				Target:   &path,
			},
			"previous":              &flags.Bool{Target: &previous},
			command.SidecarFlagName: command.SidecarFlag(&sidecar),
			"timestamps":            &flags.Bool{Target: &timestamps},
			"since":                 &flags.String{Target: &since},
			"tail":                  &flags.Int{Target: &tail},
		},
		Run: func(ctx *BuildCommandContext, args []string, flags cli.Flags) error {
			var sidecarPtr *string
			if flags[command.SidecarFlagName].Set() {
				sidecarPtr = &sidecar
			}

			return BuildLogs(
				ctx.Client,
				ctx.System,
				ctx.Build,
				path,
				sidecarPtr,
				follow,
				previous,
				timestamps,
				since,
				int64(tail),
				os.Stdout,
			)
		},
	}

	return cmd.Command()
}

// BuildLogs prints the logs for the specified build the supplied writer.
func BuildLogs(
	client client.Interface,
	system v1.SystemID,
	id v1.BuildID,
	path tree.Path,
	sidecar *string,
	follow bool,
	previous bool,
	timestamps bool,
	since string,
	tail int64,
	w io.Writer,
) error {
	options := &v1.ContainerLogOptions{
		Follow:     follow,
		Previous:   previous,
		Timestamps: timestamps,
		Since:      since,
		Tail:       &tail,
	}
	logs, err := client.V1().Systems().Builds(system).Logs(id, path, sidecar, options)
	if err != nil {
		return err
	}

	defer logs.Close()
	_, err = io.Copy(w, logs)
	return err
}
