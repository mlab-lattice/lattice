package main

import (
	"flag"

	"github.com/mlab-lattice/system/pkg/envoy/xds-api/rest"
	"github.com/mlab-lattice/system/pkg/kubernetes/envoy/xds-api/pernode"
)

var (
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.Parse()
}

func main() {
	backend, err := pernode.NewKubernetesPerNodeBackend(kubeconfig)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(backend, 8080)
}
