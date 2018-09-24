package main

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/xdg"
	"os"
)

func main() {
	configDir := xdg.ConfigDir(command.Latticectl)
	if _, err := os.Stat(fmt.Sprintf("%v/use-colons", configDir)); os.IsNotExist(err) {
		latticectl.Command.Execute()
	}
	latticectl.Command.ExecuteColon()
}
