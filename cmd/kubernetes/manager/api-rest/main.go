package main

import (
	"flag"

	"github.com/mlab-lattice/system/pkg/kubernetes/manager/backend"
	"github.com/mlab-lattice/system/pkg/manager/api/rest"
)

var (
	kubeconfig string
	port       int
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.IntVar(&port, "port", 8080, "port to bind to")
	flag.Parse()
}

func main() {
	b, err := backend.NewKubernetesBackend(kubeconfig)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(b, int32(port))
}
