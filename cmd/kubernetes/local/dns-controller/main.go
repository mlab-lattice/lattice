package main

import (
	"github.com/mlab-lattice/lattice/cmd/kubernetes/local/dns-controller/app"
)

func main() {
	app.Command().Execute()
}
