package latticectl

import (
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/latticectl/deploys"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
)

const (
	deployBuildFlag   = "build"
	deployPathFlag    = "path"
	deployVersionFlag = "version"
)

var (
	deployTypeFlags = []string{deployBuildFlag, deployPathFlag, deployVersionFlag}
)

func Deploy() *cli.Command {
	var (
		build   string
		output  string
		path    tree.Path
		version string
		watch   bool
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			deployBuildFlag: &flags.String{Target: &build},
			command.OutputFlagName: command.OutputFlag(
				&output,
				[]printer.Format{
					printer.FormatJSON,
					printer.FormatTable,
				},
				printer.FormatTable,
			),
			deployPathFlag:        &flags.Path{Target: &path},
			deployVersionFlag:     &flags.String{Target: &version},
			command.WatchFlagName: command.WatchFlag(&watch),
		},
		MutuallyExclusiveFlags: [][]string{deployTypeFlags},
		RequiredFlagSet:        [][]string{deployTypeFlags},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			format := printer.Format(output)
			switch {
			case flags[deployBuildFlag].Set():
				return DeployBuild(ctx.Client, ctx.System, v1.BuildID(build), os.Stdout, format, watch)

			case flags[deployPathFlag].Set():
				return DeployPath(ctx.Client, ctx.System, path, os.Stdout, format, watch)

			case flags[deployVersionFlag].Set():
				return DeployVersion(ctx.Client, ctx.System, v1.Version(version), os.Stdout, format, watch)

			default:
				// this shouldn't happen due to the mutually exclusive and required flag sets
				return fmt.Errorf("deploy type not set")
			}
		},
	}

	return cmd.Command()
}

func DeployBuild(
	client client.Interface,
	system v1.SystemID,
	build v1.BuildID,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	deploy, err := client.V1().Systems().Deploys(system).CreateFromBuild(build)
	if err != nil {
		return err
	}

	return displayDeploy(client, system, deploy, fmt.Sprintf("build %v", build), w, f, watch)
}

func DeployPath(
	client client.Interface,
	system v1.SystemID,
	path tree.Path,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	deploy, err := client.V1().Systems().Deploys(system).CreateFromPath(path)
	if err != nil {
		return err
	}

	return displayDeploy(client, system, deploy, fmt.Sprintf("path %v", path.String()), w, f, watch)
}

func DeployVersion(
	client client.Interface,
	system v1.SystemID,
	version v1.Version,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	deploy, err := client.V1().Systems().Deploys(system).CreateFromVersion(version)
	if err != nil {
		return err
	}

	return displayDeploy(client, system, deploy, fmt.Sprintf("version %v", version), w, f, watch)
}

func displayDeploy(
	client client.Interface,
	system v1.SystemID,
	deploy *v1.Deploy,
	description string,
	w io.Writer,
	f printer.Format,
	watch bool,
) error {
	if watch {
		return deploys.WatchDeploy(client, system, deploy.ID, w, f)
	}

	fmt.Fprintf(
		w,
		`
deploying %s for system %s. deploy ID: %s

to watch deploy, run:
    latticectl deploys status --deploy %s -w
`,
		description,
		color.IDString(string(system)),
		color.IDString(string(deploy.ID)),
		string(deploy.ID),
	)
	return nil
}
