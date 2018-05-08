package cli

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mlab-lattice/lattice/pkg/util/cli/template"
)

type Command struct {
	Name             string
	Short            string
	Args             Args
	Flags            Flags
	PreRun           func()
	Run              func(args []string)
	Subcommands      []*Command
	UsageFunc        func(*Command) error
	HelpFunc         func(*Command)
	cobraCmd         *cobra.Command
	isSpaceSeparated bool
}

func (c *Command) emptyRun(cmd *cobra.Command) {
	c.cobraCmd.SetHelpFunc(c.helpFuncWrapper)
	cmd.Help()
	os.Exit(1)
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
				c.emptyRun(cmd)
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

	c.cobraCmd.SetUsageFunc(c.usageFuncWrapper)
	c.cobraCmd.SetHelpFunc(c.helpFuncWrapper)

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

	c.isSpaceSeparated = true

	return nil
}

// usageFuncWrapper calls the correct usage function, and lets the usageFunction be called on a Command rather than a cobra.Command
func (c *Command) usageFuncWrapper(command *cobra.Command) error {
	if c.UsageFunc != nil {
		return c.UsageFunc(c)
	}

	return c.defaultUsageFunc(c)
}

// helpFuncWrapper calls the correct help function, and lets the usageFunction be called on a Command rather than a cobra.Command
func (c *Command) helpFuncWrapper(command *cobra.Command, strings []string) {
	if c.HelpFunc != nil {
		c.HelpFunc(c)
		return
	}

	c.defaultHelpFunc(c)
}

// defaultUsageFunc is the Usage function that will be called if none is provided
func (c *Command) defaultUsageFunc(command *Command) error {
	// TODO :: Seems like Usage & Help have the same use for us right now. (Perhaps just for default)
	tmplName := "defaultHelpTemplate"
	templateToExecute := "UsageTemplate"
	return template.TryExecuteTemplate(template.DefaultTemplate, tmplName, templateToExecute, template.DefaultTemplateFuncs, c)
}

// defaultHelpFunc is the Help function that will be called if none is provided
func (c *Command) defaultHelpFunc(command *Command) {
	tmplName := "defaultHelpTemplate"
	templateToExecute := "UsageTemplate"
	err := template.TryExecuteTemplate(template.DefaultTemplate, tmplName, templateToExecute, template.DefaultTemplateFuncs, c)
	if err != nil {
		log.Fatalf(err.Error())
	}
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
					c.emptyRun(cmd)
				}

				run(args)
			}
		}(subcommand.Run)

		c.cobraCmd.AddCommand(subcommand.cobraCmd)
	}

	c.isSpaceSeparated = false

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

// Template helpers

// CommandSeparator returns the string needed to invoke a subcommand. This is either a space or ':'.
func (c *Command) CommandSeparator() string {
	if c.isSpaceSeparated {
		return " "
	}

	return ":"
}

func (c *Command) IsRunnable() bool {
	return c.cobraCmd.Runnable()
}

func (c *Command) HasSubcommands() bool {
	return len(c.Subcommands) != 0
}

func (c *Command) HasFlags() bool {
	return len(c.Flags) != 0
}

func (c *Command) NamePadding() int {
	return 35
}

func (c *Command) FlagNamePadding() int {
	return 10
}

func (c *Command) CommandPath() string {
	return c.cobraCmd.CommandPath()
}

func (c *Command) FlagsSorted() Flags {
	flags := c.Flags
	sort.Sort(flags)
	return flags
}

func SortCommands(commands CommandList) CommandList {
	sort.Sort(commands)
	return commands
}

type CommandList []*Command

func (c CommandList) Len() int {
	return len(c)
}

func (c CommandList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CommandList) Less(i, j int) bool {
	return strings.Compare(c[i].CommandPath(), c[j].CommandPath()) == -1
}

// AllSubcommands returns the recursive subcommand tree as one flat sorted array.
func (c *Command) AllSubcommands() []*Command {
	// found is a list of all flattened subcommands
	found := make([]*Command, 0)
	// queue is the list of Commands that still need to be flattened
	queue := make([]*Command, 0)
	queue = c.Subcommands

	done := false
	for done == false {
		if len(queue) == 0 {
			// nothing left to search, found contains all the subcommands
			done = true
		} else {
			// explore the first element in the queue. Add this node to found and add each subcommand to the queue
			found = append(found, queue[0])
			queue = append(queue, queue[0].Subcommands...)
			queue = queue[1:]
		}
	}

	return SortCommands(found)
}

type CommandGroup struct {
	Commands  []*Command
	GroupName string
}

// SubcommandsByGroup returns commands grouped by immediate children of a node. The order is a pre order. The elements in each group are sorted.
func (c *Command) SubcommandsByGroup() []*CommandGroup {
	// found is a list of all flattened subcommands
	found := make([]*CommandGroup, 0)
	// queue is the list of Commands that still need to be flattened
	queue := make([]*Command, 0)
	queue = append(queue, c)

	done := false
	for done == false {
		if len(queue) == 0 {
			// nothing left to search, found contains all the subcommands
			done = true
		} else {
			// explore the first element in the queue. Add this node to found and add each subcommand to the queue
			nextElem := queue[0]
			queue = queue[1:]
			if len(nextElem.Subcommands) == 0 {
				continue
			}

			groupName := nextElem.CommandPath()
			newCmdGroup := &CommandGroup{
				Commands:  SortCommands(nextElem.Subcommands),
				GroupName: groupName,
			}
			found = append(found, newCmdGroup)
			queue = append(queue, nextElem.Subcommands...)
		}
	}

	return found
}
