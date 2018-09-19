package latticectl

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/builds"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

const (
	buildPathFlag    = "path"
	buildVersionFlag = "version"
)

var (
	buildTypeFlags = []string{buildPathFlag, buildVersionFlag}
)

func Build() *cli.Command {
	var (
		output  string
		path    tree.Path
		version string
		watch   bool
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			buildPathFlag:         &flags.Path{Target: &path},
			buildVersionFlag:      &flags.String{Target: &version},
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		MutuallyExclusiveFlags: [][]string{buildTypeFlags},
		RequiredFlagSet:        [][]string{buildTypeFlags},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			switch {
			case flags[buildPathFlag].Set():
				return BuildPath(ctx.Client, ctx.System, path, os.Stdout, format, watch)

			case flags[buildVersionFlag].Set():
				return BuildVersion(ctx.Client, ctx.System, v1.Version(version), os.Stdout, format, watch)

			default:
				// this shouldn't happen due to the mutually exclusive and required flag sets
				return fmt.Errorf("build type not set")
			}
		},
	}

	return cmd.Command()
}

func BuildPath(
	client client.Interface,
	system v1.SystemID,
	path tree.Path,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	build, err := client.V1().Systems().Builds(system).CreateFromPath(path)
	if err != nil {
		return err
	}

	return displayBuild(client, system, build, fmt.Sprintf("path %v", path.String()), w, f, watch)
}

func BuildVersion(
	client client.Interface,
	system v1.SystemID,
	version v1.Version,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	build, err := client.V1().Systems().Builds(system).CreateFromVersion(version)
	if err != nil {
		return err
	}

	return displayBuild(client, system, build, fmt.Sprintf("version %v", version), w, f, watch)
}

func displayBuild(
	client client.Interface,
	system v1.SystemID,
	build *v1.Build,
	description string,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	if watch {
		return builds.WatchBuild(client, system, build.ID, w, f)
	}

	fmt.Fprintf(
		w,
		`
building %s for system %s. build ID: %s

to watch build, run:
    latticectl builds status --build %s -w
`,
		description,
		color.IDString(string(system)),
		color.IDString(string(build.ID)),
		string(build.ID),
	)
	return nil
}
