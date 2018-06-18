package messages

import (
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

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
