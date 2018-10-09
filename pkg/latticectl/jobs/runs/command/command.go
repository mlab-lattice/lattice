package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	jobcommand "github.com/mlab-lattice/lattice/pkg/latticectl/jobs/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	jobRunFlagName = "job-run"
)

// JobRunCommandContext contains the information available to any JobRunCommand.
type JobRunCommandContext struct {
	*jobcommand.JobCommandContext
	JobRun v1.JobRunID
}

// JobRunCommand is a Command that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a jobRun was specified.
type JobRunCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *JobRunCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the JobRunCommand.
func (c *JobRunCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var jobRun string
	c.Flags[jobRunFlagName] = &flags.String{
		Required: true,
		Target:   &jobRun,
	}

	cmd := &jobcommand.JobCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *jobcommand.JobCommandContext, args []string, f cli.Flags) error {
			jobRunCtx := &JobRunCommandContext{
				JobCommandContext: ctx,
				JobRun:            v1.JobRunID(jobRun),
			}
			return c.Run(jobRunCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}
