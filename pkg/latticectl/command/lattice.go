package command

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

// LatticeCommandContext contains the information available to any LatticeCommand.
type LatticeCommandContext struct {
	Context *Context
	Lattice string
	Client  client.Interface
}

// LatticeCommand is a Command that acts on a specific lattice.
// More practically, it is a command that has validated that there is a valid Context for
// how to connect to a lattice and has included that information in the LatticeCommandContext.
type LatticeCommand struct {
	Short                  string
	Args                   cli.Args
	Flags                  cli.Flags
	Run                    func(ctx *LatticeCommandContext, args []string, flags cli.Flags) error
	MutuallyExclusiveFlags [][]string
	RequiredFlagSet        [][]string
	Subcommands            map[string]*cli.Command
}

// Command returns a *cli.Command for the LatticeCommand.
func (c *LatticeCommand) Command() *cli.Command {
	if c.Flags == nil {
		c.Flags = make(cli.Flags)
	}

	// allow config and context to be set via flags
	var (
		configPath  string
		contextName string
	)
	c.Flags[ConfigFlagName] = ConfigFlag(&configPath)
	c.Flags[ContextFlagName] = ContextFlag(&contextName)

	cmd := &cli.Command{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		MutuallyExclusiveFlags: c.MutuallyExclusiveFlags,
		RequiredFlagSet:        c.RequiredFlagSet,
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := ConfigFile{Path: configPath}

			// if a context wasn't explicitly set, check the config file for the current context
			if contextName == "" {
				var err error
				contextName, err = configFile.CurrentContext()
				if err != nil {
					return err
				}
			}

			// retrieve the context, bubbling up any errors such as the context not existing
			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			// create the proper client based on the Context's AuthContext
			var client client.Interface
			switch {
			case context.Auth == nil:
				client = rest.NewUnauthenticatedClient(context.URL)

			case context.Auth.BearerToken != nil:
				client = rest.NewBearerTokenClient(context.URL, *context.Auth.BearerToken)
				break

			default:
				return fmt.Errorf("invalid auth options for context %v", contextName)
			}

			ctx := &LatticeCommandContext{
				Context: context,
				Client:  client,
			}
			return c.Run(ctx, args, flags)
		},
		Subcommands: c.Subcommands,
	}

	return cmd
}
