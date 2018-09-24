package main

import (
	"github.com/mlab-lattice/lattice/cmd/backend/kubernetes/lattice-controller-manager/app"
)

func main() {
	app.Command().Execute()
}
