package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	flagNone = "none"
)

func Switch() *cli.Command {
	var (
		configPath string
		name       string
		none       bool
	)

	return &cli.Command{
		Flags: cli.Flags{
			flagName:               &flags.String{Target: &name},
			flagNone:               &flags.Bool{Target: &none},
			command.ConfigFlagName: command.ConfigFlag(&configPath),
		},
		MutuallyExclusiveFlags: [][]string{{flagName, flagNone}},
		RequiredFlagSet:        [][]string{{flagName, flagNone}},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			if none {
				return configFile.UnsetCurrentContext()
			}

			return configFile.SetCurrentContext(name)
		},
	}
}
