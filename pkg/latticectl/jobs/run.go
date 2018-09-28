package jobs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

func Run() *cli.Command {
	var (
		follow bool
		path   tree.Path
		envs   []string
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			"follow": &flags.Bool{Target: &follow},
			"env": &flags.StringArray{
				Short:  "e",
				Target: &envs,
			},
			"path": &flags.Path{
				Target:   &path,
				Required: true,
			},
		},
		Args: cli.Args{AllowAdditional: true},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			environment := make(definitionv1.ContainerExecEnvironment)
			for _, val := range envs {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format %v. expected name=val", val)
				}
			}

			if len(args) == 0 {
				args = nil
			}

			return RunJob(
				ctx.Client,
				ctx.System,
				path,
				args,
				environment,
				follow,
				os.Stdout,
			)
		},
	}

	return cmd.Command()
}

func RunJob(
	client client.Interface,
	system v1.SystemID,
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	follow bool,
	w io.Writer,
) error {
	job, err := client.V1().Systems().Jobs(system).Run(path, command, environment)
	if err != nil {
		return err
	}

	if !follow {
		return nil
	}

	return JobLogs(
		client,
		system,
		job.ID,
		nil,
		follow,
		false,
		false,
		"",
		0,
		w,
	)
}
