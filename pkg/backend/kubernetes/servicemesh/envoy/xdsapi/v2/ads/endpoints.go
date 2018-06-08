package ads

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyendpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
)

func (s *Service) getEndpoints(
	clusters []envoycache.Resource,
	systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	endpoints := make([]envoycache.Resource, 0, len(clusters))
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
			return nil, fmt.Errorf("Invalid Service path <%v>", path)
		}
		component, ok := service.Components[componentName]
		if !ok {
			return nil, fmt.Errorf("Invalid Component name <%v>", componentName)
		}
		envoyPort, ok := component.Ports[port]
		if !ok {
			return nil, fmt.Errorf("Invalid Port <%v>", port)
		}
		addresses := make([]envoyendpoint.LbEndpoint, 0, len(service.IPAddresses))
		for _, address := range service.IPAddresses {
			addresses = append(addresses, envoyendpoint.LbEndpoint{
				Endpoint: &envoyendpoint.Endpoint{
					Address: &envoycore.Address{
						Address: &envoycore.Address_SocketAddress{
							SocketAddress: &envoycore.SocketAddress{
								Protocol: envoycore.TCP,
								Address:  address,
								PortSpecifier: &envoycore.SocketAddress_PortValue{
									PortValue: uint32(envoyPort),
								},
							},
						},
					},
				},
			})
		}
		endpoints = append(endpoints, &envoyv2.ClusterLoadAssignment{
			ClusterName: cluster.EdsClusterConfig.ServiceName,
			Endpoints: []envoyendpoint.LocalityLbEndpoints{
				{
					LbEndpoints: addresses,
				},
			},
		})
	}
	return endpoints, nil
}
