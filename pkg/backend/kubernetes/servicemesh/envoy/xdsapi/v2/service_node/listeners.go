package servicenode

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoytcpproxy "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsmsgs "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node/messages"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
	// lerror "github.com/mlab-lattice/lattice/pkg/util/error"
)

func (s *ServiceNode) newHTTPEgressListener(service *xdsapi.Service) *envoyv2.Listener {
	httpFilters := []*envoyhttpcxnmgr.HttpFilter{
		xdsmsgs.NewHttpRouterFilter(),
	}

	filters := []envoylistener.Filter{
		*xdsmsgs.NewRdsHttpConnectionManagerFilter(
			xdsconstants.HTTPEgressStatPrefix, xdsconstants.RouteNameEgress, httpFilters),
	}

	filterChains := []envoylistener.FilterChain{
		*xdsmsgs.NewFilterChain(nil, nil, false, filters),
	}

	address := xdsmsgs.NewTcpSocketAddress("0.0.0.0", service.EgressPorts.HTTP)

	return xdsmsgs.NewListener(
		xdsconstants.HTTPEgressListenerName, address, filterChains)
}

func (s *ServiceNode) newTCPEgressListener(
	service *xdsapi.Service, systemServices map[tree.NodePath]*xdsapi.Service) *envoyv2.Listener {
	tcpProxyRoutes := make([]*envoytcpproxy.TcpProxy_DeprecatedV1_TCPRoute, 0, len(systemServices))
	for path, _service := range systemServices {
		for componentName, component := range _service.Components {
			for servicePort, listenerPort := range component.Ports {
				// only add routes for TCP services
				if listenerPort.Protocol != "TCP" {
					continue
				}
				clusterName := xdsutil.GetClusterNameForComponentPort(
					s.ServiceCluster(), path, componentName, servicePort)
				ips := make([]string, len(_service.IPAddresses))
				copy(ips, _service.IPAddresses)
				tcpProxyRoutes = append(tcpProxyRoutes, xdsmsgs.NewDeprecatedV1TcpProxyRoute(
					clusterName, ips, []int32{servicePort}))
			}
		}
	}

	filters := []envoylistener.Filter{
		*xdsmsgs.NewDeprecatedV1TCPProxyFilter(
			xdsconstants.TCPEgressStatPrefix, xdsmsgs.NewDeprecatedV1TcpProxyRoutes(tcpProxyRoutes)),
	}

	filterChains := []envoylistener.FilterChain{
		*xdsmsgs.NewFilterChain(nil, nil, false, filters),
	}

	address := xdsmsgs.NewTcpSocketAddress("0.0.0.0", service.EgressPorts.TCP)

	return xdsmsgs.NewOriginalDestinationListener(
		xdsconstants.TCPEgressListenerName, address, filterChains)
}

func (s *ServiceNode) newHTTPIngressListener(
	path tree.NodePath,
	listenerName, componentName string,
	servicePort, envoyPort int32) *envoyv2.Listener {
	httpFilters := []*envoyhttpcxnmgr.HttpFilter{
		xdsmsgs.NewHttpRouterFilter(),
	}

	routes := []envoyroute.Route{
		*xdsmsgs.NewRouteRoute(
			xdsmsgs.NewPrefixRouteMatch("/"),
			xdsmsgs.NewClusterRouteActionRouteRoute(
				xdsutil.GetLocalClusterNameForComponentPort(
					s.ServiceCluster(), path, componentName, servicePort))),
	}

	virtualHosts := []envoyroute.VirtualHost{
		*xdsmsgs.NewVirtualHost(
			fmt.Sprintf("%v %v port %v", path, componentName, servicePort),
			[]string{"*"},
			routes),
	}

	// FIXME: add health_check filter
	// FIXME: look into other filters (buffer, potentially add fault injection for testing)
	filters := []envoylistener.Filter{
		*xdsmsgs.NewStaticHttpConnectionManagerFilter(
			listenerName, virtualHosts, httpFilters),
	}

	filterChains := []envoylistener.FilterChain{
		*xdsmsgs.NewFilterChain(nil, nil, false, filters),
	}

	address := xdsmsgs.NewTcpSocketAddress("0.0.0.0", envoyPort)

	return xdsmsgs.NewListener(listenerName, address, filterChains)
}

func (s *ServiceNode) newTCPIngressListener(
	path tree.NodePath,
	listenerName, componentName string,
	servicePort, envoyPort int32) *envoyv2.Listener {
	filters := []envoylistener.Filter{
		*xdsmsgs.NewTCPProxyFilter(
			listenerName, xdsutil.GetLocalClusterNameForComponentPort(
				s.ServiceCluster(), path, componentName, servicePort)),
	}

	filterChains := []envoylistener.FilterChain{
		*xdsmsgs.NewFilterChain(nil, nil, false, filters),
	}

	address := xdsmsgs.NewTcpSocketAddress("0.0.0.0", envoyPort)

	return xdsmsgs.NewListener(listenerName, address, filterChains)
}

func (s *ServiceNode) getListeners(
	systemServices map[tree.NodePath]*xdsapi.Service) (listeners []envoycache.Resource, err error) {
	// TODO <GEB>: lerror not working as expected here: [unable to retrieve function/file/line]
	// NOTE: https://github.com/golang/go/wiki/PanicAndRecover#usage-in-a-package
	//       support nested builder funcs
	// defer func() {
	// 	if _panic := recover(); _panic != nil {
	// 		err = lerror.Errorf("%v", _panic)
	// 	}
	// }()

	listeners = make([]envoycache.Resource, 0)

	path, err := s.Path()
	if err != nil {
		return nil, err
	}

	service, ok := systemServices[path]
	if !ok {
		return nil, fmt.Errorf("invalid Service path: %v", path)
	}

	listeners = append(listeners, s.newHTTPEgressListener(service))
	listeners = append(listeners, s.newTCPEgressListener(service, systemServices))

	// There's a listener for each port of Service, listening on the port's EnvoyPort
	for componentName, component := range service.Components {
		for port, listenerPort := range component.Ports {
			var listener *envoyv2.Listener
			listenerName := fmt.Sprintf(
				"%v %v port %v %v ingress", path, componentName, port, listenerPort.Protocol)

			switch listenerPort.Protocol {
			case "HTTP":
				listener = s.newHTTPIngressListener(
					path, listenerName, componentName, port, listenerPort.Port)
			case "TCP":
				listener = s.newTCPIngressListener(
					path, listenerName, componentName, port, listenerPort.Port)
			default:
				return nil, fmt.Errorf("invalid Service protocol: %v", listenerPort.Protocol)
			}

			listeners = append(listeners, listener)
		}
	}
	return listeners, nil
}
