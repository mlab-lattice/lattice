package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Command struct {
	Name        string
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string)
	Subcommands []*Command
	cobraCmd    *cobra.Command
	UsageFunc 	func(*cobra.Command) error
	HelpFunc  	func(*cobra.Command) error
}

func (c *Command) Execute() {
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *Command) Init() error {
	c.cobraCmd = &cobra.Command{
		Use:   c.Name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}
			c.Run(args)
		},
	}

	if err := c.addArgs(); err != nil {
		return fmt.Errorf("error adding args: %v", err)
	}

	if err := c.addFlags(); err != nil {
		return fmt.Errorf("error adding flags: %v", err)
	}

	if err := c.addSubcommands(); err != nil {
		return fmt.Errorf("error initializing subcommands: %v", err)
	}

	if c.UsageFunc != nil {
		c.cobraCmd.SetUsageFunc(c.UsageFunc)
	}

	if c.HelpFunc != nil {
		c.cobraCmd.SetUsageFunc(c.HelpFunc)
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

func (c *Command) addFlags() error {
	names := make(map[string]struct{})
	for _, flag := range c.Flags {
		if err := flag.Validate(); err != nil {
			return fmt.Errorf("error validating flag %v: %v", flag.GetName(), err)
		}

		if _, ok := names[flag.GetName()]; ok {
			return fmt.Errorf("multiple flags with the name %v", flag.GetName())
		}

		flag.AddToFlagSet(c.cobraCmd.Flags())
		names[flag.GetName()] = struct{}{}
	}

	return nil
}

func (c *Command) getFlagParsers() map[string]func() error {
	parsers := make(map[string]func() error)
	for _, flag := range c.Flags {
		parser := flag.Parse()
		if parser != nil {
			parsers[flag.GetName()] = parser
		}
	}

	return parsers
}

func (c *Command) addSubcommands() error {
	names := make(map[string]struct{})
	for _, subcommand := range c.Subcommands {
		if _, ok := names[subcommand.Name]; ok {
			return fmt.Errorf("multiple subcommands with the name %v", c.Name)
		}

		if err := subcommand.Init(); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", c.Name, err)
		}

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
		names[subcommand.Name] = struct{}{}
	}

	return nil
}

func (c *Command) Help() {
	c.cobraCmd.Help()
}

func (c *Command) Usage() {
	c.cobraCmd.Usage()
}

func (c *Command) ExecuteColon() {
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	if err := c.initColon(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *Command) initColon() error {
	c.cobraCmd = &cobra.Command{
		Use:   c.Name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}

			c.Run(args)
		},
	}

	for _, subcommand := range c.Subcommands {
		if err := subcommand.Init(); err != nil {
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
		subcommand.cobraCmd.Run = func(run func([]string)) func(*cobra.Command, []string) {
			return func(cmd *cobra.Command, args []string) {
				if run == nil {
					cmd.Help()
					os.Exit(1)
				}

				run(args)
			}
		}(subcommand.Run)

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
	}

	return nil
}

func (c *Command) getSubcommands2(path string, subcommands []*Command) ([]*Command, error) {
	var ret []*Command
	for _, subcommand := range subcommands {
		name := fmt.Sprintf("%v%v", path, subcommand.Name)
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
