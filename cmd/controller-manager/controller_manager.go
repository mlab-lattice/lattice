package controller_manager

import (
	"flag"

	"github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
}

func main() {
	app.Run(kubeconfig)
}
