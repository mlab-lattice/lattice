package main

import (
	"github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app"
)

func main() {
	app.Command().Execute()
}
