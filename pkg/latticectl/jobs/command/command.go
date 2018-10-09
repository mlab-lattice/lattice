package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	jobFlagName = "job"
)

// JobCommandContext contains the information available to any JobCommand.
type JobCommandContext struct {
	*command.SystemCommandContext
	Job v1.JobID
}

// JobCommand is a Command that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a job was specified.
type JobCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *JobCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the JobCommand.
func (c *JobCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var job string
	c.Flags[jobFlagName] = &flags.String{
		Required: true,
		Target:   &job,
	}

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			jobCtx := &JobCommandContext{
				SystemCommandContext: ctx,
				Job:                  v1.JobID(job),
			}
			return c.Run(jobCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}
