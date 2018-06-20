package main

import (
	"flag"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/util"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/backend/pernode"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/grpc"
)

var (
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.Parse()
}

func main() {
	stopCh := util.SetupSignalHandler()

	backend, err := pernode.NewKubernetesPerNodeBackend(kubeconfig, stopCh)
	if err != nil {
		panic(err)
	}

	grpc.RunNewGRPCServer(backend, 8080, stopCh)
}
