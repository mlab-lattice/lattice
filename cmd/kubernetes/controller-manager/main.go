package main

import (
	"github.com/mlab-lattice/lattice/cmd/kubernetes/controller-manager/app"
)

func main() {
	app.Command().Execute()
}
