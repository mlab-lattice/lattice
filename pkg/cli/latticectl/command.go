package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
)

type Command interface {
	Base() (*BaseCommand, error)
}

type BaseCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(latticectl *Latticectl, args []string)
	Subcommands []Command
}

func (c *BaseCommand) Base() (*BaseCommand, error) {
	return c, nil
}

func (c *BaseCommand) Command(latticectl *Latticectl) (*command.Command, error) {
	var subcommands []*command.Command
	for _, subcmd := range c.Subcommands {
		base, err := subcmd.Base()
		if err != nil {
			return nil, err
		}

		cmd, err := base.Command(latticectl)
		if err != nil {
			return nil, err
		}

		subcommands = append(subcommands, cmd)
	}

	cmd := &command.Command{
		Name:        c.Name,
		Short:       c.Short,
		Args:        c.Args,
		Flags:       c.Flags,
		Subcommands: subcommands,
	}
	if c.Run != nil {
		cmd.Run = func(args []string) {
			c.Run(latticectl, args)
		}
	}

	return cmd, nil
}
