package context

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

func Update() *cli.Command {
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
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			flagBearerToken:        &flags.String{Target: &bearerToken},
			flagName:               &flags.String{Target: &name},
			command.SystemFlagName: command.SystemFlag(&system),
			flagUnauthenticated:    &flags.Bool{Target: &unauthenticated},
			flagURL:                &flags.String{Target: &url},
		},
		MutuallyExclusiveFlags: [][]string{authFlags},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			if !flags[flagName].Set() {
				var err error
				name, err = configFile.CurrentContext()
				if err != nil {
					return err
				}
			}

			context, err := configFile.Context(name)
			if err != nil {
				return err
			}

			switch {
			case bearerToken != "":
				context.Auth = &command.AuthContext{BearerToken: &bearerToken}

			case unauthenticated:
				context.Auth = &command.AuthContext{}
			}

			// if changing the URL, unset the system as well
			if flags[flagURL].Set() {
				context.URL = url
				context.System = ""
			}

			if flags[command.SystemFlagName].Set() {
				context.System = v1.SystemID(system)
			}

			return configFile.UpdateContext(name, context)
		},
	}
}
