package servicenode

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyendpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsmsgs "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node/messages"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
	lerror "github.com/mlab-lattice/lattice/pkg/util/error"
)

func (s *ServiceNode) getEndpoints(
	clusters []envoycache.Resource,
	systemServices map[tree.NodePath]*xdsapi.Service) (endpoints []envoycache.Resource, err error) {
	// NOTE: https://github.com/golang/go/wiki/PanicAndRecover#usage-in-a-package
	//       support nested builder funcs
	defer func() {
		if _panic := recover(); _panic != nil {
			err = lerror.Errorf("%v", _panic)
		}
	}()

	endpoints = make([]envoycache.Resource, 0, len(clusters))

	for _, resource := range clusters {
		cluster := resource.(*envoyv2.Cluster)
		if cluster.EdsClusterConfig == nil {
			continue
		}
		_, path, componentName, port, err :=
			xdsutil.GetPartsFromClusterName(cluster.EdsClusterConfig.ServiceName)
		if err != nil {
			return nil, err
		}
		service, ok := systemServices[path]
		if !ok {
			return nil, fmt.Errorf("invalid Service path <%v>", path)
		}
		component, ok := service.Components[componentName]
		if !ok {
			return nil, fmt.Errorf("invalid Component name <%v>", componentName)
		}
		listenerPort, ok := component.Ports[port]
		if !ok {
			return nil, fmt.Errorf("invalid Port <%v>", port)
		}
		addresses := make([]envoyendpoint.LbEndpoint, 0, len(service.IPAddresses))
		for _, address := range service.IPAddresses {
			addresses = append(
				addresses, *xdsmsgs.NewLbEndpoint(
					xdsmsgs.NewTcpSocketAddress(address, listenerPort.Port)))
		}
		endpoints = append(
			endpoints, xdsmsgs.NewClusterLoadAssignment(
				cluster.EdsClusterConfig.ServiceName, addresses))
	}
	return endpoints, err
}
