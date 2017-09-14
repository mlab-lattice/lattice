package main

import (
	"flag"

	"github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app"
	"github.com/mlab-lattice/kubernetes-integration/pkg/provider"
)

var (
	kubeconfig string
	p          provider.Interface
)

func init() {
	var providerName string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&providerName, "provider", "", "name of provider")
	flag.Parse()

	p = provider.GetProvider(providerName)
}

func main() {
	app.Run(kubeconfig, p)
}
