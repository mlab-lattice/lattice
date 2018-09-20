package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

func Delete() *cli.Command {
	var (
		configPath string
		name       string
	)

	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			flagName: &flags.String{
				Required: true,
				Target:   &name,
			},
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}
			return configFile.DeleteContext(name)
		},
	}
}
