package main

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/xdg"
)

func main() {
	configDir := xdg.ConfigDir(command.Latticectl)
	if _, err := os.Stat(fmt.Sprintf("%v/use-colons", configDir)); os.IsNotExist(err) {
		Command().Execute()
	}
	Command().ExecuteColon()
}

func Command() *cli.RootCommand {
	return &latticectl.Command
}
