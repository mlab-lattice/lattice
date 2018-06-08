package service_node

import (
	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
)

func (s *ServiceNode) getClusters(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	clusters := make([]envoycache.Resource, 0)

	for path, service := range systemServices {
		servicePath, err := s.Path()
		if err != nil {
			return nil, err
		}
		isLocalService := servicePath == path

		for componentName, component := range service.Components {
			for port := range component.Ports {
				clusterName := xdsutil.GetClusterNameForComponentPort(
					s.ServiceCluster(), path, componentName, port)
				clusters = append(clusters, &envoyv2.Cluster{
					Name: clusterName,
					Type: envoyv2.Cluster_EDS,
					// TODO: figure out a good value for this
					ConnectTimeout: xdsconstants.ClusterConnectTimeout,
					LbPolicy:       envoyv2.Cluster_ROUND_ROBIN,
					EdsClusterConfig: &envoyv2.Cluster_EdsClusterConfig{
						EdsConfig: &envoycore.ConfigSource{
							ConfigSourceSpecifier: &envoycore.ConfigSource_Ads{
								Ads: &envoycore.AggregatedConfigSource{},
							},
						},
						ServiceName: clusterName,
					},
				})

				if isLocalService {
					clusterName = xdsutil.GetLocalClusterNameForComponentPort(
						s.ServiceCluster(), path, componentName, port)
					clusters = append(clusters, &envoyv2.Cluster{
						Name: clusterName,
						Type: envoyv2.Cluster_STATIC,
						// TODO: figure out a good value for this
						ConnectTimeout: xdsconstants.ClusterConnectTimeout,
						LbPolicy:       envoyv2.Cluster_ROUND_ROBIN,
						Hosts: []*envoycore.Address{
							{
								Address: &envoycore.Address_SocketAddress{
									SocketAddress: &envoycore.SocketAddress{
										Protocol: envoycore.TCP,
										Address:  "127.0.0.1",
										PortSpecifier: &envoycore.SocketAddress_PortValue{
											PortValue: uint32(port),
										},
									},
								},
							},
						},
					})
				}
			}
		}
	}

	return clusters, nil
}
