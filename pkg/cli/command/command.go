package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var OnInitialize = cobra.OnInitialize

type Command interface {
	Execute()
	Init() error
	name() string
	cobraCommand() *cobra.Command
}

type BasicCommand struct {
	Name        string
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string)
	Subcommands []Command
	cobraCmd    *cobra.Command
}

func (c *BasicCommand) Execute() {
	if err := c.validate(); err != nil {
		c.exit(err)
	}

	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *BasicCommand) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name must be set")
	}

	return nil
}

func (c *BasicCommand) Init() error {
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

	return nil
}

func (c *BasicCommand) addArgs() error {
	if err := c.Args.validate(); err != nil {
		return err
	}

	for _, arg := range c.Args {
		c.cobraCmd.Use += fmt.Sprintf(" [%v]", arg.Name)
	}

	c.cobraCmd.Args = cobra.MinimumNArgs(c.Args.min())

	return nil
}

func (c *BasicCommand) addFlags() error {
	names := make(map[string]struct{})
	for _, flag := range c.Flags {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("error validating flag %v: %v", flag.name(), err)
		}

		if _, ok := names[flag.name()]; ok {
			return fmt.Errorf("multiple flags with the name %v", flag.name())
		}

		flag.addToCmd(c.cobraCmd)
		names[flag.name()] = struct{}{}
	}

	return nil
}

func (c *BasicCommand) addSubcommands() error {
	names := make(map[string]struct{})
	for _, subcommand := range c.Subcommands {
		if err := subcommand.Init(); err != nil {
			return err
		}

		if _, ok := names[subcommand.name()]; ok {
			return fmt.Errorf("multiple subcommands with the name %v", c.Name)
		}

		if err := subcommand.Init(); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", c.Name, err)
		}

		c.cobraCmd.AddCommand(subcommand.cobraCommand())
		names[subcommand.name()] = struct{}{}
	}

	return nil
}

func (c *BasicCommand) name() string {
	return c.Name
}

func (c *BasicCommand) cobraCommand() *cobra.Command {
	return c.cobraCmd
}

//func (c *BasicCommand) ExecuteColon() {
//	if err := c.validate(); err != nil {
//		c.exit(err)
//	}
//
//	cmd, err := c.initColon()
//	if err != nil {
//		c.exit(err)
//	}
//
//	c.exit(cmd.Execute())
//}
//
//func (c *BasicCommand) initColon() (*cobra.Command, error) {
//	root := &cobra.Command{
//		Use:   c.Name,
//		Short: c.Short,
//		Run: func(cmd *cobra.Command, args []string) {
//			if c.Run == nil {
//				cmd.Help()
//				os.Exit(1)
//			}
//
//			c.Run(args)
//		},
//	}
//
//	for _, subcommand := range c.getSubcommands("") {
//		// why does this need to be an immediately invoked function?
//		// answer here: https://www.ardanlabs.com/blog/2014/06/pitfalls-with-closures-in-go.html
//		// (n.b. subcommand.Name will be copied here since it's a string, but since
//		//  subcommand.Run is a pointer, we need to do this trickery)
//		cmd := func(run func([]string)) *cobra.Command {
//			return &cobra.Command{
//				Use: subcommand.Name,
//				Run: func(cmd *cobra.Command, args []string) {
//					run(args)
//				},
//			}
//		}(subcommand.Run)
//
//		if err := subcommand.addArgs(); err != nil {
//			return nil, fmt.Errorf("error adding args: %v", err)
//		}
//
//		if err := subcommand.addFlags(); err != nil {
//			return nil, fmt.Errorf("error adding flags: %v", err)
//		}
//
//		root.AddCommand(cmd)
//	}
//
//	return root, nil
//}
//
//func (c *BasicCommand) getSubcommands(path string) []*BasicCommand {
//	var subcommands []*BasicCommand
//	for _, subcommand := range c.Subcommands {
//		subcommand.name() = fmt.Sprintf("%v%v", path, subcommand.name())
//		subcommands = append(subcommands, subcommand)
//		for _, subsubcommand := range subcommand.getSubcommands(fmt.Sprintf("%v:", subcommand.Name)) {
//			subcommands = append(subcommands, subsubcommand)
//		}
//
//	}
//
//	return subcommands
//}

func (c *BasicCommand) exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
