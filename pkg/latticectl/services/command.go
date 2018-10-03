package services

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

const (
	serviceFlag     = "service"
	servicePathFlag = "path"
)

var serviceIdentifierFlags = []string{serviceFlag, servicePathFlag}

// ServiceCommandContext contains the information available to any ServiceCommand.
type ServiceCommandContext struct {
	*command.SystemCommandContext
	Service v1.ServiceID

	service *v1.Service
}

// ServiceCommand is a Command that acts on a specific build in a specific system.
// More practically, it is a valid SystemCommand and also validates that a service was specified.
type ServiceCommand struct {
	Name                   string
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *ServiceCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command

	ctx *ServiceCommandContext
}

// Command returns a *cli.Command for the ServiceCommand.
func (c *ServiceCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	var (
		service string
		path    tree.Path
	)
	c.Flags[serviceFlag] = &flags.String{Target: &service}
	c.Flags[servicePathFlag] = &flags.Path{Target: &path}

	c.MutuallyExclusiveFlags = append(c.MutuallyExclusiveFlags, serviceIdentifierFlags)
	c.RequiredFlagSet = append(c.MutuallyExclusiveFlags, serviceIdentifierFlags)

	cmd := &command.SystemCommand{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(ctx *command.SystemCommandContext, args []string, f cli.Flags) error {
			serviceCtx := &ServiceCommandContext{
				SystemCommandContext: ctx,
			}
			switch {
			case f[serviceFlag].Set():
				serviceCtx.Service = v1.ServiceID(service)

			case f[servicePathFlag].Set():
				service, err := ctx.Client.V1().Systems().Services(ctx.System).GetByPath(path)
				if err != nil {
					return err
				}

				serviceCtx.Service = service.ID
				serviceCtx.service = service
			}

			c.ctx = serviceCtx
			return c.Run(serviceCtx, args, f)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Command()
}
