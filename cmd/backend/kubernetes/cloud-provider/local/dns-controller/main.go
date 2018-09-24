package main

import (
	"github.com/mlab-lattice/lattice/cmd/backend/kubernetes/cloud-provider/local/dns-controller/app"
)

func main() {
	app.Command().Execute()
}
