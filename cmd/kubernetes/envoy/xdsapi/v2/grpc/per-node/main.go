package main

import (
	"flag"
	"net"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/util"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/backend/pernode"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/grpc"
)

var (
	kubeconfig        string
	redirectCIDRBlock string
)

// FIXME(kevindrosendahl): convert this to pkg/util/cli
func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	// XXX <GEB>: should we be using cli here?
	flag.StringVar(&redirectCIDRBlock, "redirect-cidr-block", "", "overlay network CIDR block")
	flag.Parse()
}

func main() {
	_, net, err := net.ParseCIDR(redirectCIDRBlock)
	if err != nil {
		panic(err)
	}

	stopCh := util.SetupSignalHandler()

	backend, err := pernode.NewKubernetesPerNodeBackend(kubeconfig, net, stopCh)
	if err != nil {
		panic(err)
	}

	grpc.RunNewGRPCServer(backend, 8080, stopCh)
}
