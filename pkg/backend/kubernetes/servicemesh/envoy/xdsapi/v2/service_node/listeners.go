package servicenode

import (
	"fmt"

	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsmsgs "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node/messages"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
	lerror "github.com/mlab-lattice/lattice/pkg/util/error"
)

func (s *ServiceNode) getListeners(systemServices map[tree.Path]*xdsapi.Service) (listeners []envoycache.Resource, err error) {
	// NOTE: https://github.com/golang/go/wiki/PanicAndRecover#usage-in-a-package
	//       support nested builder funcs
	defer func() {
		if _panic := recover(); _panic != nil {
			err = lerror.Errorf("%v", _panic)
		}
	}()

	listeners = make([]envoycache.Resource, 0)

	path, err := s.Path()
	if err != nil {
		return nil, err
	}

	service, ok := systemServices[path]
	if !ok {
		return nil, fmt.Errorf("invalid Service path <%v>", path)
	}

	httpFilters := []*envoyhttpcxnmgr.HttpFilter{
		xdsmsgs.NewHttpRouterFilter(),
	}

	filters := []envoylistener.Filter{
		*xdsmsgs.NewRdsHttpConnectionManagerFilter(
			"egress", xdsconstants.RouteNameEgress, httpFilters),
	}

	filterChains := []envoylistener.FilterChain{
		*xdsmsgs.NewFilterChain(nil, nil, false, filters),
	}

	address := xdsmsgs.NewTcpSocketAddress("0.0.0.0", service.EgressPort)

	listeners = append(listeners, xdsmsgs.NewListener("egress", address, filterChains))

	// There's a listener for each port of Service, listening on the port's EnvoyPort
	for componentName, component := range service.Components {
		for port, envoyPort := range component.Ports {
			listenerName := fmt.Sprintf("%v %v port %v ingress", path, componentName, port)

			httpFilters = []*envoyhttpcxnmgr.HttpFilter{
				xdsmsgs.NewHttpRouterFilter(),
			}

			routes := []envoyroute.Route{
				*xdsmsgs.NewRouteRoute(
					xdsmsgs.NewPrefixRouteMatch("/"),
					xdsmsgs.NewClusterRouteActionRouteRoute(
						xdsutil.GetLocalClusterNameForComponentPort(
							s.ServiceCluster(), path, componentName, port))),
			}

			virtualHosts := []envoyroute.VirtualHost{
				*xdsmsgs.NewVirtualHost(
					fmt.Sprintf("%v %v port %v", path, componentName, port),
					[]string{"*"},
					routes),
			}

			// FIXME: add health_check filter
			// FIXME: look into other filters (buffer, potentially add fault injection for testing)
			filters = []envoylistener.Filter{
				*xdsmsgs.NewStaticHttpConnectionManagerFilter(
					listenerName, virtualHosts, httpFilters),
			}

			filterChains = []envoylistener.FilterChain{
				*xdsmsgs.NewFilterChain(nil, nil, false, filters),
			}

			address = xdsmsgs.NewTcpSocketAddress("0.0.0.0", envoyPort)

			listeners = append(listeners, xdsmsgs.NewListener(listenerName, address, filterChains))
		}
	}
	return listeners, nil
}
