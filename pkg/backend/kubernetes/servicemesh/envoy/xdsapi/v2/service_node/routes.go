package service_node

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
)

func (s *ServiceNode) getRoutes(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	route := &envoyv2.RouteConfiguration{
		Name:         xdsconstants.RouteNameEgress,
		VirtualHosts: []envoyroute.VirtualHost{},
	}

	for path, service := range systemServices {
		for componentName, component := range service.Components {
			for port := range component.Ports {
				domain := fmt.Sprintf("%v.local", path.ToDomain())
				domains := []string{fmt.Sprintf("%v:%v", domain, port)}

				// Should be able to access an HTTP component on port 80 via either:
				//   - http://path.to.service:80
				//   - http://path.to.service
				if port == xdsconstants.PortHTTPDefault {
					domains = append(domains, domain)
				}

				route.VirtualHosts = append(route.VirtualHosts, envoyroute.VirtualHost{
					Name:    string(path),
					Domains: domains,
					Routes: []envoyroute.Route{
						{
							Match: envoyroute.RouteMatch{
								PathSpecifier: &envoyroute.RouteMatch_Prefix{
									Prefix: "/",
								},
							},
							Action: &envoyroute.Route_Route{
								Route: &envoyroute.RouteAction{
									ClusterSpecifier: &envoyroute.RouteAction_Cluster{
										Cluster: xdsutil.GetClusterNameForComponentPort(s.ServiceCluster(), path, componentName, port),
									},
								},
							},
						},
					},
				})
			}
		}
	}

	return []envoycache.Resource{route}, nil
}
