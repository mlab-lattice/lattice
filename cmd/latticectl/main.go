package main

import "github.com/mlab-lattice/lattice/cmd/latticectl/definition"

func main() {
	latticectl := definition.GenerateLatticeCtl()
	latticectl.ExecuteColon()
}
