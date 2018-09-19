package secrets

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	secretFlagName = "secret"
)

type Command struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *SecretCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

type SecretCommandContext struct {
	*command.SystemCommandContext
	Secret tree.PathSubcomponent
}

func (c *Command) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var secret tree.PathSubcomponent
	c.Flags[secretFlagName] = &flags.PathSubcomponent{
		Required: true,
		Target:   &secret,
	}

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			secretCtx := &SecretCommandContext{
				SystemCommandContext: ctx,
				Secret:               secret,
			}
			return c.Run(secretCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}
