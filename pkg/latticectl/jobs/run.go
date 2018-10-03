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

const (
	numRetriesFlag = "num-retries"
)

// Run returns a *cli.Command to run a job.
func Run() *cli.Command {
	var (
		envs       []string
		follow     bool
		numRetries int32
		path       tree.Path
		secrets    []string
	)

	cmd := command.SystemCommand{
		Flags: map[string]cli.Flag{
			"follow": &flags.Bool{Target: &follow},
			"env": &flags.StringArray{
				Short:  "e",
				Target: &envs,
			},
			numRetriesFlag: &flags.Int32{Target: &numRetries},
			"path": &flags.Path{
				Target:   &path,
				Required: true,
			},
			"secret": &flags.StringArray{Target: &secrets},
		},
		Args: cli.Args{AllowAdditional: true},
		Run: func(ctx *command.SystemCommandContext, args []string, flags cli.Flags) error {
			environment := make(definitionv1.ContainerExecEnvironment)
			for _, val := range envs {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format %v. expected name=val", val)
				}

				value := parts[1]
				environment[parts[0]] = definitionv1.ValueOrSecret{Value: &value}
			}

			for _, val := range secrets {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					return fmt.Errorf("invalid secret format %v. expected name=/path/to:secret", val)
				}

				key := parts[0]
				if _, ok := environment[key]; ok {
					return fmt.Errorf("key %v set as both plaintext environment variable and secret", key)
				}

				secret, err := tree.NewPathSubcomponent(parts[1])
				if err != nil {
					return fmt.Errorf("invalid secret format %v. expected name=/path/to:secret", val)
				}

				environment[key] = definitionv1.ValueOrSecret{
					SecretRef: &definitionv1.SecretRef{Value: secret},
				}
			}

			if len(args) == 0 {
				args = nil
			}

			var numRetriesPtr *int32
			if flags[numRetriesFlag].Set() {
				numRetriesPtr = &numRetries
			}

			return RunJob(
				ctx.Client,
				ctx.System,
				path,
				args,
				environment,
				numRetriesPtr,
				follow,
				os.Stdout,
			)
		},
	}

	return cmd.Command()
}

// RunJob runs the supplied job.
func RunJob(
	client client.Interface,
	system v1.SystemID,
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	numRetries *int32,
	follow bool,
	w io.Writer,
) error {
	job, err := client.V1().Systems().Jobs(system).Run(path, command, environment, numRetries)
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
