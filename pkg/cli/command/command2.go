package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"log"
)

type Command2 interface {
	BaseCommand() (*BaseCommand2, error)
}

func Execute(c Command2) {
	cmd, err := c.BaseCommand()
	if err != nil {
		log.Panic(err)
	}
	cmd.Execute()
}

func ExecuteColon(c Command2) {
	cmd, err := c.BaseCommand()
	if err != nil {
		log.Panic(err)
	}
	cmd.ExecuteColon()
}

type BaseCommand2 struct {
	Name        string
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string)
	Subcommands []Command2
	cobraCmd    *cobra.Command
	subcommands map[Command2]*BaseCommand2
}

func (c *BaseCommand2) BaseCommand() (*BaseCommand2, error) {
	return c, nil
}

func (c *BaseCommand2) mergeSubcommands(cmds map[Command2]*BaseCommand2) {
	for cmd, bc := range cmds {
		if _, ok := c.subcommands[cmd]; ok {
			continue
		}

		c.subcommands[cmd] = bc
	}
}

func (c *BaseCommand2) getSubcommand(cmd Command2) (*BaseCommand2, error) {
	if bc, ok := c.subcommands[cmd]; ok {
		return bc, nil
	}

	bc, err := cmd.BaseCommand()
	if err != nil {
		return nil, err
	}

	c.subcommands[cmd] = bc
	return bc, nil
}

func (c *BaseCommand2) Execute() {
	if err := c.validate(); err != nil {
		c.exit(err)
	}

	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *BaseCommand2) Init() error {
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

func (c *BaseCommand2) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name must be set")
	}

	return nil
}

func (c *BaseCommand2) addArgs() error {
	if err := c.Args.validate(); err != nil {
		return err
	}

	for _, arg := range c.Args {
		c.cobraCmd.Use += fmt.Sprintf(" [%v]", arg.Name)
	}

	c.cobraCmd.Args = cobra.MinimumNArgs(c.Args.min())

	return nil
}

func (c *BaseCommand2) addFlags() error {
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

func (c *BaseCommand2) addSubcommands() error {
	names := make(map[string]struct{})
	for _, subcommand := range c.Subcommands {
		//cmd, err := subcommand.BaseCommand()
		cmd, err := c.getSubcommand(subcommand)
		if err != nil {
			return err
		}

		if _, ok := names[cmd.Name]; ok {
			return fmt.Errorf("multiple subcommands with the name %v", c.Name)
		}

		cmd.subcommands = c.subcommands
		if err := cmd.Init(); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", c.Name, err)
		}
		c.mergeSubcommands(cmd.subcommands)

		c.cobraCmd.AddCommand(cmd.cobraCmd)
		names[cmd.Name] = struct{}{}
	}

	return nil
}

func (c *BaseCommand2) ExecuteColon() {
	c.subcommands = make(map[Command2]*BaseCommand2)
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	if err := c.initColon(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *BaseCommand2) initColon() error {
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
		//cmd, err := subcommand.BaseCommand()
		cmd, err := c.getSubcommand(subcommand)
		if err != nil {
			return err
		}

		if err := cmd.Init(); err != nil {
			return err
		}
	}

	subcommands, err := c.getSubcommands2("", c.Subcommands)
	if err != nil {
		return err
	}

	for _, subcommand := range subcommands {
		//cmd, err := subcommand.BaseCommand()
		cmd, err := c.getSubcommand(subcommand)
		if err != nil {
			return err
		}

		// why does this need to be an immediately invoked function?
		// answer here: https://www.ardanlabs.com/blog/2014/06/pitfalls-with-closures-in-go.html
		// (n.b. subcommand.Name will be copied here since it's a string, but since
		//  subcommand.Run is a pointer, we need to do this trickery)
		cmd.cobraCmd.Run = func(run func([]string)) func(*cobra.Command, []string) {
			return func(cmd *cobra.Command, args []string) {
				if run == nil {
					cmd.Help()
					os.Exit(1)
				}

				run(args)
			}
		}(cmd.Run)

		c.cobraCmd.AddCommand(cmd.cobraCmd)
	}

	return nil
}

func (c *BaseCommand2) getSubcommands2(path string, subcommands []Command2) ([]Command2, error) {
	var ret []Command2
	for _, subcommand := range subcommands {
		//cmd, err := subcommand.BaseCommand()
		cmd, err := c.getSubcommand(subcommand)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("%v%v", path, cmd.Name)
		cmd.cobraCmd.Use = name
		ret = append(ret, subcommand)

		subsubcommands, err := c.getSubcommands2(fmt.Sprintf("%v:", name), cmd.Subcommands)
		if err != nil {
			return nil, err
		}

		for _, subsubcommand := range subsubcommands {
			ret = append(ret, subsubcommand)
		}

	}

	return ret, nil
}

func (c *BaseCommand2) exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
