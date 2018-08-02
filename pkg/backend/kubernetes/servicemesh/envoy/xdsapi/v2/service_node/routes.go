package servicenode

import (
	"fmt"

	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsmsgs "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node/messages"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
	lerror "github.com/mlab-lattice/lattice/pkg/util/error"
)

func (s *ServiceNode) getRoutes(
	systemServices map[tree.Path]*xdsapi.Service) (routes []envoycache.Resource, err error) {
	// NOTE: https://github.com/golang/go/wiki/PanicAndRecover#usage-in-a-package
	//       support nested builder funcs
	defer func() {
		if _panic := recover(); _panic != nil {
			err = lerror.Errorf("%v", _panic)
		}
	}()

	virtualHosts := make([]envoyroute.VirtualHost, 0)

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
				virtualHosts = append(
					virtualHosts, *xdsmsgs.NewVirtualHost(
						string(path), domains, []envoyroute.Route{
							*xdsmsgs.NewRouteRoute(
								xdsmsgs.NewPrefixRouteMatch("/"),
								xdsmsgs.NewClusterRouteActionRouteRoute(
									xdsutil.GetClusterNameForComponentPort(
										s.ServiceCluster(), path, componentName, port))),
						}))
			}
		}
	}

	return []envoycache.Resource{
		xdsmsgs.NewRouteConfiguration(xdsconstants.RouteNameEgress, virtualHosts),
	}, err
}
