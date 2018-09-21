package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

type RootCommand struct {
	Name string
	*Command
}

type Command struct {
	Short                  string
	Args                   Args
	Flags                  Flags
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	PreRun                 func()
	Run                    func(args []string, flags Flags) error
	Subcommands            map[string]*Command
	cobraCmd               *cobra.Command
}

func (c *RootCommand) Execute() {
	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *RootCommand) Init() error {
	return c.Command.init(c.Name)
}

func (c *Command) init(name string) error {
	if err := c.initCommand(name); err != nil {
		return err
	}
	return c.initSubcommands()
}

func (c *Command) initCommand(name string) error {
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

	c.cobraCmd.PreRun = func(cmd *cobra.Command, args []string) {
		c.checkMutexFlags(cmd)
		c.checkRequiredFlagSets(cmd)

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

func (c *Command) initSubcommands() error {
	for name, subcommand := range c.Subcommands {
		if err := subcommand.init(name); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", name, err)
		}

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
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

func (c *Command) checkMutexFlags(cmd *cobra.Command) {
	// check to see if any of the flags are mutually exclusive
	// first build up a set of flags that each flag conflicts with
	conflictFlags := make(map[string][]string)
	for _, mutexSet := range c.MutuallyExclusiveFlags {
		for _, flag := range mutexSet {
			conflicts, ok := conflictFlags[flag]
			if !ok {
				conflicts = make([]string, 0)
			}

			for _, conflict := range mutexSet {
				// don't add self to our own conflict set
				if conflict == flag {
					continue
				}
				conflicts = append(conflicts, conflict)
			}
			conflictFlags[flag] = conflicts
		}
	}

	// then go through each flag _that was set_. if we have already seen flags that conflict
	// with this flag, set a conflict message
	// otherwise, for each other flag that the flag conflicts with, create or add to a list of
	// flags that conflict with the other flag. then go on to the next flag.
	// this means that if the next flag conflicted with the current flag, it will see that the
	// current flag is listed under flags that conflict with it, and it would set the conflict
	// message
	conflict := ""
	pendingConflict := make(map[string][]string)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// bail early if we already found a conflict
		if conflict != "" {
			return
		}

		// if this flag triggers a conflict, set the conflict string and return
		pendingConflicts, ok := pendingConflict[flag.Name]
		if ok {
			conflict = fmt.Sprintf("flag %v is mutually exclusive with %v", flag.Name, strings.Join(pendingConflicts, ", "))
			return
		}

		conflicts, ok := conflictFlags[flag.Name]
		if ok {
			for _, conflict := range conflicts {
				existing, ok := pendingConflict[conflict]
				if !ok {
					existing = make([]string, 0)
				}

				pendingConflict[conflict] = append(existing, flag.Name)
			}
		}
	})
	if conflict != "" {
		fmt.Printf("%v\n", conflict)
		os.Exit(1)
	}
}

func (c *Command) checkRequiredFlagSets(cmd *cobra.Command) {
	// get all flags and which requirement sets seeing each flag
	// would fulfil
	membership := make(map[string][]int)
	unfulfilled := make(map[int]struct{})
	for idx, requiredSets := range c.RequiredFlagSet {
		unfulfilled[idx] = struct{}{}
		for _, required := range requiredSets {
			members, ok := membership[required]
			if !ok {
				members = make([]int, 0)
			}

			membership[required] = append(members, idx)
		}
	}

	// for each flag, remove the indexes of its memberships from unfulfilled
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		members, ok := membership[flag.Name]
		if !ok {
			return
		}

		for _, idx := range members {
			delete(unfulfilled, idx)
		}
	})

	// if there's unfulfilled requirement sets, print one of them
	for idx := range unfulfilled {
		fmt.Printf("must set one of %v\n", strings.Join(c.RequiredFlagSet[idx], ", "))
		os.Exit(1)
	}
}

func (c *Command) Help() {
	c.cobraCmd.Help()
}

func (c *RootCommand) ExecuteColon() {
	if err := c.InitColon(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *RootCommand) InitColon() error {
	if err := c.initCommand(c.Name); err != nil {
		return err
	}

	return c.Command.initColonSubcommands(c, "")
}

// initColonSubcommands currently just adds the subcommands to the root subcommand
// this means that only <root-cmd> -h will show the commands
// if you have foo and foo:bar, <root-cmd> -h will not show foo:bar as a subcommand
// if we want to invest in colon commands this should probably be fixed through help
// templates
func (c *Command) initColonSubcommands(root *RootCommand, path string) error {
	for name, subcommand := range c.Subcommands {
		if path != "" {
			name = fmt.Sprintf("%v:%v", path, name)
		}

		if err := subcommand.initCommand(name); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", name, err)
		}

		root.cobraCmd.AddCommand(subcommand.cobraCmd)

		if err := subcommand.initColonSubcommands(root, name); err != nil {
			return err
		}
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
