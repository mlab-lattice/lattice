package context

//
//import (
//	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
//	"github.com/mlab-lattice/lattice/pkg/util/cli2"
//	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
//)
//
//const (
//	switchFlagName = "name"
//)
//
//func Switch() *cli.Command {
//	return &cli.Command{
//		Flags: cli.Flags{
//			switchFlagName:         &flags.String{Required: true},
//			command.ConfigFlagName: command.ConfigFlag(),
//		},
//		Run: func(args []string, flags cli.Flags) error {
//			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
//			configPath := flags[command.ConfigFlagName].Value().(string)
//			configFile := command.ConfigFile{Path: configPath}
//
//			contextName := flags[switchFlagName].Value().(string)
//			return configFile.SetCurrentContext(contextName)
//		},
//	}
//}
