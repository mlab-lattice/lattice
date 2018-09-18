package context

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

func Update() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			flagBearerToken:        &flags.String{},
			flagName:               &flags.String{},
			flagUnauthenticated:    &flags.Bool{},
			flagURL:                &flags.String{},
			command.ConfigFlagName: command.ConfigFlag(),
			command.SystemFlagName: command.SystemFlag(),
		},
		MutuallyExclusiveFlags: [][]string{authFlags},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			var contextName string
			if flags[flagName].Set() {
				contextName = flags[flagName].Value().(string)
			} else {
				var err error
				contextName, err = configFile.CurrentContext()
				if err != nil {
					return err
				}
			}

			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			bearerToken := flags[flagBearerToken].Value().(string)
			if bearerToken != "" {
				context.Auth = &command.AuthContext{BearerToken: &bearerToken}
			}

			if flags[flagUnauthenticated].Value().(bool) {
				context.Auth = &command.AuthContext{}
			}

			// if changing the URL, unset the system as well
			if flags[flagURL].Set() {
				context.URL = flags[flagURL].Value().(string)
				context.System = ""
			}

			if flags[command.SystemFlagName].Set() {
				context.System = v1.SystemID(flags[command.SystemFlagName].Value().(string))
			}

			return configFile.UpdateContext(contextName, context)
		},
	}
}
