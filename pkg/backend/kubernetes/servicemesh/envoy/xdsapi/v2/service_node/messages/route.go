package messages

import (
	"fmt"
	"strings"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoytcpproxy "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
)

// ------------------------------
// HTTP connection manager routes
// ------------------------------

func NewClusterRouteActionRouteRoute(
	clusterName string) *envoyroute.Route_Route {
	return &envoyroute.Route_Route{
		Route: &envoyroute.RouteAction{
			ClusterSpecifier: &envoyroute.RouteAction_Cluster{
				Cluster: clusterName,
			},
		},
	}
}

func NewPrefixRouteMatch(prefix string) *envoyroute.RouteMatch {
	return &envoyroute.RouteMatch{
		PathSpecifier: &envoyroute.RouteMatch_Prefix{
			Prefix: prefix,
		},
	}
}

func NewRouteRoute(
	match *envoyroute.RouteMatch, action *envoyroute.Route_Route) *envoyroute.Route {
	return &envoyroute.Route{
		Match:  *match,
		Action: action,
	}
}

func NewVirtualHost(
	name string, domains []string, routes []envoyroute.Route) *envoyroute.VirtualHost {
	return &envoyroute.VirtualHost{
		Name:    name,
		Domains: domains,
		Routes:  routes,
	}
}

func NewRouteConfiguration(
	name string, virtualHosts []envoyroute.VirtualHost) *envoyv2.RouteConfiguration {
	return &envoyv2.RouteConfiguration{
		Name:         name,
		VirtualHosts: virtualHosts,
	}
}

// ----------------
// TCP proxy routes
// ----------------

func NewDeprecatedV1TcpProxyRoutes(
	routes []*envoytcpproxy.TcpProxy_DeprecatedV1_TCPRoute) *envoytcpproxy.TcpProxy_DeprecatedV1 {
	return &envoytcpproxy.TcpProxy_DeprecatedV1{
		Routes: routes,
	}
}

func NewDeprecatedV1TcpProxyRoute(
	cluster string,
	destinationIPs []string,
	destinationPorts []int32) *envoytcpproxy.TcpProxy_DeprecatedV1_TCPRoute {
	destinationIPList := make([]*envoycore.CidrRange, len(destinationIPs))
	for ip := range destinationIPs {
		destinationIPList = append(destinationIPList, &envoycore.CidrRange{
			AddressPrefix: ip,
			PrefixLen: pbtypes.UInt32Value{
				Value: 32,
			},
		})
	}
	destinationPortList := strings.Trim(
		strings.Join(fmt.Sprint(destinationPorts), ","), "[]")
	return &envoytcpproxy.TcpProxy_DeprecatedV1_TCPRoute{
		Cluster:           cluster,
		DestinationIpList: destinationIPList,
		DestinationPorts:  destinationPortList,
	}
}
