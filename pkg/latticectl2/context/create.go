package context

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	flagBearerToken     = "bearer-token"
	flagUnauthenticated = "unauthenticated"
	flagName            = "name"
	flagURL             = "url"
)

var authFlags = []string{flagBearerToken, flagUnauthenticated}

func Create() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			flagBearerToken:        &flags.String{},
			flagName:               &flags.String{Required: true},
			flagUnauthenticated:    &flags.Bool{},
			flagURL:                &flags.String{Required: true},
			command.ConfigFlagName: command.ConfigFlag(),
			command.SystemFlagName: command.SystemFlag(),
		},
		MutuallyExclusiveFlags: [][]string{authFlags},
		RequiredFlagSet:        [][]string{authFlags},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			contextName := flags[flagName].Value().(string)

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
			bearerToken := flags[flagBearerToken].Value().(string)
			if bearerToken != "" {
				auth = &command.AuthContext{BearerToken: &bearerToken}
			}

			context := &command.Context{
				URL:    flags[flagURL].Value().(string),
				System: v1.SystemID(flags[command.SystemFlagName].Value().(string)),
				Auth:   auth,
			}
			return configFile.CreateContext(contextName, context)
		},
	}
}
