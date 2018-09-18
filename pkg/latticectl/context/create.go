package context

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	flagBearerToken     = "bearer-token"
	flagName            = "name"
	flagUnauthenticated = "unauthenticated"
	flagURL             = "url"
)

var authFlags = []string{flagBearerToken, flagUnauthenticated}

func Create() *cli.Command {
	var (
		bearerToken     string
		configPath      string
		name            string
		system          string
		unauthenticated bool
		url             string
	)

	return &cli.Command{
		Flags: cli.Flags{
			flagBearerToken:        &flags.String{Target: &bearerToken},
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			flagName: &flags.String{
				Required: true,
				Target:   &name,
			},
			command.SystemFlagName: command.SystemFlag(&system),
			flagUnauthenticated:    &flags.Bool{Target: &unauthenticated},
			flagURL: &flags.String{
				Required: true,
				Target:   &url,
			},
		},
		MutuallyExclusiveFlags: [][]string{authFlags},
		RequiredFlagSet:        [][]string{authFlags},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			// if the context already exists, can't create it
			_, err := configFile.Context(name)
			if err == nil {
				return command.NewContextAlreadyExistsError(name)
			}

			// if the error was something other than the context not existing,
			// bubble it up
			if _, ok := err.(*command.InvalidContextError); !ok {
				return err
			}

			var auth *command.AuthContext
			if bearerToken != "" {
				auth = &command.AuthContext{BearerToken: &bearerToken}
			}

			context := &command.Context{
				URL:    url,
				System: v1.SystemID(system),
				Auth:   auth,
			}
			return configFile.CreateContext(name, context)
		},
	}
}
