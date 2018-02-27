package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Command struct {
	Name        string
	Args        Args
	Flags       []Flag
	Run         func(args []string)
	Subcommands []Command
	path        []string
}

func (c *Command) Execute() {
	if err := c.validate(); err != nil {
		c.exit(err)
	}

	cmd, err := c.init()
	if err != nil {
		c.exit(err)
	}

	c.exit(cmd.Execute())
}

func (c *Command) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name must be set")
	}

	return nil
}

func (c *Command) init() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: c.Name,
		Run: func(cmd *cobra.Command, args []string) {
			c.Run(args)
		},
	}

	if err := c.addArgs(cmd); err != nil {
		return nil, fmt.Errorf("error adding args: %v", err)
	}

	if err := c.addFlags(cmd); err != nil {
		return nil, fmt.Errorf("error adding flags: %v", err)
	}

	if err := c.addSubcommands(cmd); err != nil {
		return nil, fmt.Errorf("error initializing subcommands: %v", err)
	}

	return cmd, nil
}

func (c *Command) addArgs(cmd *cobra.Command) error {
	if err := c.Args.validate(); err != nil {
		return err
	}

	for _, arg := range c.Args {
		cmd.Use += fmt.Sprintf(" [%v]", arg.Name)
	}

	cmd.Args = cobra.MinimumNArgs(c.Args.min())

	return nil
}

func (c *Command) addFlags(cmd *cobra.Command) error {
	names := make(map[string]struct{})
	for _, flag := range c.Flags {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("error validating flag %v: %v", flag.name(), err)
		}

		if _, ok := names[flag.name()]; ok {
			return fmt.Errorf("multiple flags with the name %v", flag.name())
		}

		flag.addToCmd(cmd)
		names[flag.name()] = struct{}{}
	}

	return nil
}

func (c *Command) addSubcommands(cmd *cobra.Command) error {
	names := make(map[string]struct{})
	for _, subcommand := range c.Subcommands {
		if err := subcommand.validate(); err != nil {
			return err
		}

		if _, ok := names[subcommand.Name]; ok {
			return fmt.Errorf("multiple subcommands with the name %v", c.Name)
		}

		subCmd, err := subcommand.init()
		if err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", c.Name, err)
		}

		cmd.AddCommand(subCmd)
		names[subcommand.Name] = struct{}{}
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
