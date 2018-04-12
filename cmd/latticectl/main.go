package main

import "github.com/mlab-lattice/lattice/cmd/latticectl/app"

func main() {
	latticectl := app.Latticectl
	latticectl.ExecuteColon()
}
