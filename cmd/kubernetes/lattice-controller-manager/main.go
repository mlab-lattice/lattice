package main

import (
	"flag"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app"
)

var (
	kubeconfig string
	provider   string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.Parse()
}

func main() {
	app.Run(kubeconfig, provider)
}
