package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type JobCommand struct {
	Name        string
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx JobCommandContext, args []string)
	Subcommands []Command
}

type JobCommandContext interface {
	SystemCommandContext
	JobID() v1.JobID
}

type jobCommandContext struct {
	SystemCommandContext
	jobID v1.JobID
}

func (c *jobCommandContext) JobID() v1.JobID {
	return c.jobID
}

func (c *JobCommand) Base() (*BaseCommand, error) {
	var jobID string
	jobIDFlag := &cli.StringFlag{
		Name:     "job",
		Required: true,
		Target:   &jobID,
	}

	flags := append(c.Flags, jobIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &jobCommandContext{
				SystemCommandContext: sctx,
				jobID:                v1.JobID(jobID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}
