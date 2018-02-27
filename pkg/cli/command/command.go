package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Command struct {
	Flags       map[string]Flag
	Run         func(args []string)
	Subcommands map[string]Command
}

func (c *Command) Execute() {
	cmd, err := c.init()
	if err != nil {
		c.exit(err)
	}

	c.exit(cmd.Execute())
}

func (c *Command) init() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			c.Run(args)
		},
	}

	if err := c.addFlags(cmd); err != nil {
		return nil, err
	}

	if err := c.addSubcommands(cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (c *Command) addFlags(cmd *cobra.Command) error {
	for name, f := range c.Flags {
		if err := f.validate(); err != nil {
			return fmt.Errorf("error validating flag %v: %v", name, err)
		}
		f.addToCmd(cmd, name)
	}

	return nil
}

func (c *Command) addSubcommands(cmd *cobra.Command) error {
	for name, subcommand := range c.Subcommands {
		subCmd, err := subcommand.init()
		if err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", name, err)
		}

		subCmd.Use = name
		cmd.AddCommand(subCmd)
	}

	return nil
}

func (c *Command) exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
