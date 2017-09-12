package main

import (
	"flag"
	"fmt"

	providerutil "github.com/mlab-lattice/core/pkg/provider"
	"github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app"
)

var (
	kubeconfig string
	provider   string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&provider, "provider", "", "path to kubeconfig file")
	flag.Parse()

	if !providerutil.ValidateProvider(provider) {
		panic(fmt.Sprintf("Invalid provider %v", provider))
	}
}

func main() {
	app.Run(kubeconfig, provider)
}
