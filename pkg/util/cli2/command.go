package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RootCommand struct {
	Name string
	*Command
}

type Command struct {
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string, flags Flags) error
	Subcommands map[string]*Command
	cobraCmd    *cobra.Command
}

func (c *RootCommand) Execute() {
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *RootCommand) Init() error {
	return c.Command.Init(c.Name)
}

func (c *Command) Init(name string) error {
	c.cobraCmd = &cobra.Command{
		Use:   name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}

			if err := c.Run(args, c.Flags); err != nil {
				c.exit(err)
			}
		},
	}

	if err := c.addArgs(); err != nil {
		return fmt.Errorf("error adding args: %v", err)
	}

	c.addFlags()

	if err := c.addSubcommands(); err != nil {
		return fmt.Errorf("error initializing subcommands: %v", err)
	}

	c.cobraCmd.PreRun = func(cmd *cobra.Command, args []string) {
		for name, parser := range c.getFlagParsers() {
			err := parser()
			if err != nil {
				fmt.Printf("error parsing flag %v: %v\n", name, err)
				os.Exit(1)
			}
		}

		if c.PreRun != nil {
			c.PreRun()
		}
	}

	return nil
}

func (c *Command) addArgs() error {
	if err := c.Args.validate(); err != nil {
		return err
	}

	for _, arg := range c.Args {
		c.cobraCmd.Use += fmt.Sprintf(" [%v]", arg.Name)
	}

	min, max := c.Args.num()
	c.cobraCmd.Args = cobra.RangeArgs(min, max)
	if min == max {
		c.cobraCmd.Args = cobra.ExactArgs(min)
	}

	return nil
}

func (c *Command) addFlags() {
	for name, flag := range c.Flags {
		flag.AddToFlagSet(name, c.cobraCmd.Flags())
	}
}

func (c *Command) getFlagParsers() map[string]func() error {
	parsers := make(map[string]func() error)
	for name, flag := range c.Flags {
		parser := flag.Parse()
		if parser != nil {
			parsers[name] = parser
		}
	}

	return parsers
}

func (c *Command) addSubcommands() error {
	for name, subcommand := range c.Subcommands {
		if err := subcommand.Init(name); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", name, err)
		}

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
	}

	return nil
}

func (c *Command) Help() {
	c.cobraCmd.Help()
}

func (c *RootCommand) ExecuteColon() {
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	if err := c.initColon(c.Name); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *Command) initColon(name string) error {
	c.cobraCmd = &cobra.Command{
		Use:   name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}

			c.Run(args, c.Flags)
		},
	}

	for _, subcommand := range c.Subcommands {
		if err := subcommand.Init(name); err != nil {
			return err
		}
	}

	subcommands, err := c.getSubcommands2("", c.Subcommands)
	if err != nil {
		return err
	}

	for _, subcommand := range subcommands {
		// why does this need to be an immediately invoked function?
		// answer here: https://www.ardanlabs.com/blog/2014/06/pitfalls-with-closures-in-go.html
		// (n.b. subcommand.Name will be copied here since it's a string, but since
		//  subcommand.Run is a pointer, we need to do this trickery)
		subcommand.cobraCmd.Run = func(run func([]string, Flags) error) func(*cobra.Command, []string) {
			return func(cmd *cobra.Command, args []string) {
				if run == nil {
					cmd.Help()
					os.Exit(1)
				}

				if err := run(args, subcommand.Flags); err != nil {
					c.exit(err)
				}
			}
		}(subcommand.Run)

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
	}

	return nil
}

func (c *Command) getSubcommands2(path string, subcommands map[string]*Command) ([]*Command, error) {
	var ret []*Command
	for name, subcommand := range subcommands {
		name := fmt.Sprintf("%v%v", path, name)
		subcommand.cobraCmd.Use = name
		ret = append(ret, subcommand)

		subsubcommands, err := c.getSubcommands2(fmt.Sprintf("%v:", name), subcommand.Subcommands)
		if err != nil {
			return nil, err
		}

		for _, subsubcommand := range subsubcommands {
			ret = append(ret, subsubcommand)
		}

	}

	return ret, nil
}

func (c *Command) exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
