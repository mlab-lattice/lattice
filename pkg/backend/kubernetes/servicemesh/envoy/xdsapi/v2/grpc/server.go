package grpc

import (
	"fmt"
	"net"

	"github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/golang/glog"
	"google.golang.org/grpc"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
)

func RunNewGRPCServer(b xdsapi.Backend, port int32, stopCh <-chan struct{}) {
	glog.Infof("Per-node GRPC server starting on port %d", port)

	xdsServer := server.NewServer(b.XDSCache(), b)
	grpcServer := grpc.NewServer()
	v2.RegisterAggregatedDiscoveryServiceServer(grpcServer, xdsServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	go b.Run(2)

	if err = grpcServer.Serve(lis); err != nil {
		glog.Error(err)
	}

	glog.Info("Per-node GRPC server waiting for stop signal")

	<-stopCh

	glog.Info("Per-node GRPC server stopped")
}
