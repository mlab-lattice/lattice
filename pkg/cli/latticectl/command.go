package latticectl

import (
	"fmt"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

type LatticeClientGenerator func(lattice string) client.Interface

type LatticeCommandContext interface {
	Lattice() string
	Client() client.Interface
}

type latticeCommandContext struct {
	lattice       string
	latticeClient client.Interface
}

func (c *latticeCommandContext) Lattice() string {
	return c.lattice
}

func (c *latticeCommandContext) Client() client.Interface {
	return c.latticeClient
}

func DefaultLatticeClient(lattice string) client.Interface {
	return rest.NewClient(lattice)
}

type Command interface {
	command.Command2
	setClient(generator LatticeClientGenerator, init bool) error
	setContext(manager ContextManager, init bool) error
}

type BaseCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx ContextManager, clientGenerator LatticeClientGenerator)
	Client      LatticeClientGenerator
	Context     ContextManager
	Subcommands []command.Command2
}

func (c *BaseCommand) BaseCommand() (*command.BaseCommand2, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	if err := c.initClient(); err != nil {
		return nil, err
	}

	cmd := &command.BaseCommand2{
		Name:        c.Name,
		Short:       c.Short,
		Args:        c.Args,
		Flags:       c.Flags,
		PreRun:      c.PreRun,
		Subcommands: c.Subcommands,
	}

	if c.Run != nil {
		cmd.Run = func(args []string) {
			c.Run(args, c.Context, c.Client)
		}
	}

	return cmd, nil
}

func (c *BaseCommand) initClient() error {
	if c.Client == nil {
		return nil
	}

	return c.setClient(c.Client, true)
}

func (c *BaseCommand) setClient(clientFunc LatticeClientGenerator, init bool) error {
	// if my client func has already been set, and this isn't an init pass,
	// then I've also already set all my subcommands' client funcs. nothing else to do
	if c.Client != nil && !init {
		return nil
	}

	c.Client = clientFunc

	// otherwise, offer the client up to my subcommands.
	// if they have their own clients set already then the'll just
	// decline the client via the above guard. if they don't, they'll
	// accept it pass it down to their children.
	// this should result in all of the subcommands inheriting the
	// client func closest to them in the tree
	for _, subcommand := range c.Subcommands {
		cmd, ok := subcommand.(Command)
		if !ok {
			continue
		}

		if err := cmd.setClient(clientFunc, false); err != nil {
			return err
		}
	}

	return nil
}

func (c *BaseCommand) initContext() error {
	if c.Context == nil {
		return nil
	}

	return c.setContext(c.Context, true)
}

func (c *BaseCommand) setContext(contextManager ContextManager, init bool) error {
	// if my context manager has already been set, and this isn't an init pass,
	// then I've also already set all my subcommands' context managers. nothing else to do
	if c.Context != nil && !init {
		return nil
	}

	fmt.Printf("setting context %#v for %v\n", contextManager, c.Name)
	c.Context = contextManager

	// otherwise, offer the client up to my subcommands.
	// if they have their own clients set already then the'll just
	// decline the client via the above guard. if they don't, they'll
	// accept it pass it down to their children.
	// this should result in all of the subcommands inheriting the
	// client func closest to them in the tree
	for _, subcommand := range c.Subcommands {
		cmd, ok := subcommand.(Command)
		if !ok {
			continue
		}

		if err := cmd.setContext(contextManager, false); err != nil {
			return err
		}
	}

	return nil
}

//
//type Command interface {
//	command.Command
//	setClient(generator LatticeClientGenerator, init bool) error
//	setContext(manager ContextManager, init bool) error
//}
//
//type BaseCommand struct {
//	Name        string
//	Short       string
//	Args        command.Args
//	Flags       command.Flags
//	PreRun      func()
//	Run         func(args []string)
//	Client      LatticeClientGenerator
//	Context     ContextManager
//	Subcommands []Command
//	*command.BaseCommand
//}
//
//func (c *BaseCommand) ExecuteColon() {
//	c.Init()
//	c.BaseCommand.ExecuteColon()
//}
//
//func (c *BaseCommand) Init() error {
//	var subcommands []command.Command
//	for _, subcommand := range c.Subcommands {
//		subcommands = append(subcommands, subcommand)
//	}
//
//	c.BaseCommand = &command.BaseCommand{
//		Name:        c.Name,
//		Short:       c.Short,
//		Args:        c.Args,
//		Flags:       c.Flags,
//		PreRun:      c.PreRun,
//		Run:         c.Run,
//		Subcommands: subcommands,
//	}
//
//	if err := c.BaseCommand.Init(); err != nil {
//		return err
//	}
//
//	if err := c.initClient(); err != nil {
//		return err
//	}
//
//	return c.initContext()
//}
//
//func (c *BaseCommand) initClient() error {
//	if c.Client == nil {
//		return nil
//	}
//
//	return c.setClient(c.Client, true)
//}
//
//func (c *BaseCommand) setClient(clientFunc LatticeClientGenerator, init bool) error {
//	// if my client func has already been set, and this isn't an init pass,
//	// then I've also already set all my subcommands' client funcs. nothing else to do
//	if c.Client != nil && !init {
//		return nil
//	}
//
//	c.Client = clientFunc
//
//	// otherwise, offer the client up to my subcommands.
//	// if they have their own clients set already then the'll just
//	// decline the client via the above guard. if they don't, they'll
//	// accept it pass it down to their children.
//	// this should result in all of the subcommands inheriting the
//	// client func closest to them in the tree
//	for _, subcommand := range c.Subcommands {
//		if err := subcommand.setClient(clientFunc, false); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (c *BaseCommand) initContext() error {
//	if c.Context == nil {
//		return nil
//	}
//
//	return c.setContext(c.Context, true)
//}
//
//func (c *BaseCommand) setContext(contextManager ContextManager, init bool) error {
//	// if my context manager has already been set, and this isn't an init pass,
//	// then I've also already set all my subcommands' context managers. nothing else to do
//	if c.Context != nil && !init {
//		return nil
//	}
//
//	fmt.Printf("setting context %#v for %v\n", contextManager, c.Name)
//	c.Context = contextManager
//
//	// otherwise, offer the client up to my subcommands.
//	// if they have their own clients set already then the'll just
//	// decline the client via the above guard. if they don't, they'll
//	// accept it pass it down to their children.
//	// this should result in all of the subcommands inheriting the
//	// client func closest to them in the tree
//	for _, subcommand := range c.Subcommands {
//		if err := subcommand.setContext(contextManager, false); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
