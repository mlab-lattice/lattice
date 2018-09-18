package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	flagNone = "none"
)

func Switch() *cli.Command {
	return &cli.Command{
		Flags: cli.Flags{
			flagName:               &flags.String{},
			flagNone:               &flags.Bool{},
			command.ConfigFlagName: command.ConfigFlag(),
		},
		MutuallyExclusiveFlags: [][]string{{flagName, flagNone}},
		RequiredFlagSet:        [][]string{{flagName, flagNone}},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := flags[command.ConfigFlagName].Value().(string)
			configFile := command.ConfigFile{Path: configPath}

			if flags[flagNone].Value().(bool) {
				return configFile.UnsetCurrentContext()
			}

			contextName := flags[flagName].Value().(string)
			return configFile.SetCurrentContext(contextName)
		},
	}
}
