package main

import (
	"github.com/mlab-lattice/lattice/cmd/backend/kubernetes/api-server/rest/app"
)

func main() {
	app.Command().Execute()
}
