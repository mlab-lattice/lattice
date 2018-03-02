package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type BuildCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	PreRun      func()
	Run         func(args []string, ctx BuildCommandContext)
	Subcommands []command.Command2
	*command.BaseCommand2
}

type BuildCommandContext interface {
	SystemCommandContext
	BuildID() types.SystemBuildID
}

type buildCommandContext struct {
	SystemCommandContext
	buildID types.SystemBuildID
}

func (c *buildCommandContext) BuildID() types.SystemBuildID {
	return c.buildID
}

func (c *BuildCommand) BaseCommand() (*command.BaseCommand2, error) {
	var buildID string
	buildIDFlag := &command.StringFlag{
		Name:     "build",
		Required: true,
		Target:   &buildID,
	}
	flags := append(c.Flags, buildIDFlag)

	cmd := &SystemCommand{
		Name:   c.Name,
		Short:  c.Short,
		Args:   c.Args,
		Flags:  flags,
		PreRun: c.PreRun,
		Run: func(args []string, sctx SystemCommandContext) {
			ctx := &buildCommandContext{
				SystemCommandContext: sctx,
				buildID:              types.SystemBuildID(buildID),
			}
			c.Run(args, ctx)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.BaseCommand()
}

//type BuildCommand struct {
//	Name        string
//	Short       string
//	Args        command.Args
//	Flags       command.Flags
//	PreRun      func()
//	Run         func(args []string, ctx BuildCommandContext)
//	Subcommands []Command
//	*SystemCommand
//}
//
//type BuildCommandContext interface {
//	SystemCommandContext
//	BuildID() types.SystemBuildID
//}
//
//type buildCommandContext struct {
//	SystemCommandContext
//	buildID types.SystemBuildID
//}
//
//func (c *buildCommandContext) BuildID() types.SystemBuildID {
//	return c.buildID
//}
//
//func (c *BuildCommand) Init() error {
//	var buildID string
//	buildIDFlag := &command.StringFlag{
//		Name:     "build",
//		Required: true,
//		Target:   &buildID,
//	}
//	flags := append(c.Flags, buildIDFlag)
//
//	c.SystemCommand = &SystemCommand{
//		Name:   c.Name,
//		Short:  c.Short,
//		Args:   c.Args,
//		Flags:  flags,
//		PreRun: c.PreRun,
//		Run: func(args []string, sctx SystemCommandContext) {
//			ctx := &buildCommandContext{
//				SystemCommandContext: sctx,
//				buildID:              types.SystemBuildID(buildID),
//			}
//			c.Run(args, ctx)
//		},
//		Subcommands: c.Subcommands,
//	}
//
//	return c.SystemCommand.Init()
//}
