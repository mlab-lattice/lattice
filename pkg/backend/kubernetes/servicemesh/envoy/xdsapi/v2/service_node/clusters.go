package service_node

import (
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsmsgs "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node/messages"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
	lerror "github.com/mlab-lattice/lattice/pkg/util/error"
)

func (s *ServiceNode) getClusters(
	systemServices map[tree.NodePath]*xdsapi.Service) (clusters []envoycache.Resource, err error) {
	defer func() {
		if _panic := recover(); _panic != nil {
			err = lerror.Errorf("%v", _panic)
		}
	}()

	clusters = make([]envoycache.Resource, 0)

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

				clusters = append(clusters, xdsmsgs.NewEdsCluster(
					clusterName,
					xdsconstants.ClusterConnectTimeout,
					xdsconstants.ClusterLbPolicyRoundRobin))

				if isLocalService {
					clusterName = xdsutil.GetLocalClusterNameForComponentPort(
						s.ServiceCluster(), path, componentName, port)

					clusters = append(clusters, xdsmsgs.NewStaticCluster(
						clusterName,
						xdsconstants.ClusterConnectTimeout,
						xdsconstants.ClusterLbPolicyRoundRobin,
						[]*envoycore.Address{xdsmsgs.NewTcpSocketAddress(xdsconstants.Localhost, port)}))
				}
			}
		}
	}

	return clusters, err
}
