package main

import (
	"flag"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/backend/pernode"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/rest"
)

var (
	kubeconfig string
	namespace  string
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
