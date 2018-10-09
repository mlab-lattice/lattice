package command

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	secretFlagName = "secret"
)

// SecretCommandContext contains the information available to any SecretCommand.
type SecretCommandContext struct {
	*command.SystemCommandContext
	Secret tree.PathSubcomponent
}

// SecretCommand is a SecretCommand that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a secret was specified.
type SecretCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *SecretCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the SecretCommand.
func (c *SecretCommand) Command() *cli.Command {
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
