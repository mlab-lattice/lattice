package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	createFlagBearerToken     = "bearer-token"
	createFlagUnauthenticated = "unauthenticated"
	createFlagName            = "name"
	createFlagURL             = "url"
)

func Create() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			createFlagBearerToken:     &flags.String{},
			createFlagName:            &flags.String{Required: true},
			createFlagUnauthenticated: &flags.Bool{},
			createFlagURL:             &flags.String{Required: true},
			command.ConfigFlagName:    command.ConfigFlag(),
		},
		MutuallyExclusiveFlags: [][]string{
			{createFlagBearerToken, createFlagUnauthenticated},
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			contextName := flags[createFlagName].Value().(string)

			// if the context already exists, can't create it
			_, err := configFile.Context(contextName)
			if err == nil {
				return command.NewContextAlreadyExistsError(contextName)
			}

			// if the error was something other than the context not existing,
			// bubble it up
			if _, ok := err.(*command.InvalidContextError); !ok {
				return err
			}

			var auth *command.AuthContext
			bearerToken := flags[createFlagBearerToken].Value().(string)
			if bearerToken != "" {
				auth = &command.AuthContext{BearerToken: &bearerToken}
			}

			context := command.Context{
				Lattice: flags[createFlagURL].Value().(string),
				Auth:    auth,
			}
			err = configFile.CreateContext(contextName, context)
			if err != nil {
				return err
			}

			return configFile.SetCurrentContext(contextName)
		},
	}
}
