package latticectl

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/builds"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

const (
	buildPathFlag    = "path"
	buildVersionFlag = "version"
)

var (
	buildTypeFlags = []string{buildPathFlag, buildVersionFlag}
)

// Build returns a *cli.Command to build a system.
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
				return BuildPath(ctx.Client, ctx.System, path, format, watch)

			case flags[buildVersionFlag].Set():
				return BuildVersion(ctx.Client, ctx.System, v1.Version(version), format, watch)

			default:
				// this shouldn't happen due to the mutually exclusive and required flag sets
				return fmt.Errorf("build type not set")
			}
		},
	}

	return cmd.Command()
}

// BuildPath builds a system with the supplied path.
func BuildPath(
	client client.Interface,
	system v1.SystemID,
	path tree.Path,
	f printer.Format,
	watch bool,
) error {
	build, err := client.V1().Systems().Builds(system).CreateFromPath(path)
	if err != nil {
		return err
	}

	return displayBuild(client, system, build, fmt.Sprintf("path %v", path.String()), f, watch)
}

// BuildPath builds a system with the supplied version.
func BuildVersion(
	client client.Interface,
	system v1.SystemID,
	version v1.Version,
	f printer.Format,
	watch bool,
) error {
	build, err := client.V1().Systems().Builds(system).CreateFromVersion(version)
	if err != nil {
		return err
	}

	return displayBuild(client, system, build, fmt.Sprintf("version %v", version), f, watch)
}

func displayBuild(
	client client.Interface,
	system v1.SystemID,
	build *v1.Build,
	description string,
	f printer.Format,
	watch bool,
) error {
	if watch {
		return builds.WatchBuildStatus(client, system, build.ID, f)
	}

	fmt.Printf(
		`
building %s for system %s. build ID: %s

to watch build, run:
    latticectl builds status --system %s --build %s -w
`,
		description,
		color.IDString(string(system)),
		color.IDString(string(build.ID)),
		system,
		build.ID,
	)
	return nil
}
