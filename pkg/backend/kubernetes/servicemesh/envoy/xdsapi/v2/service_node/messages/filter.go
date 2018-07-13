package messages

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "strings"

	"github.com/golang/glog"

	pbjson "github.com/gogo/protobuf/jsonpb"
	pbtypes "github.com/gogo/protobuf/types"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyhttprouter "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/router/v2"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoytcpproxy "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoyutil "github.com/envoyproxy/go-control-plane/pkg/util"

	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
)

// -------------
// filter chains
// -------------

func NewFilterChain(
	filterChainMatch *envoylistener.FilterChainMatch,
	tlsContext *envoyauth.DownstreamTlsContext,
	useProxyProto bool,
	filters []envoylistener.Filter) *envoylistener.FilterChain {
	return &envoylistener.FilterChain{
		FilterChainMatch: filterChainMatch,
		TlsContext:       tlsContext,
		UseProxyProto:    &pbtypes.BoolValue{Value: useProxyProto},
		Filters:          filters,
	}
}

// ------------
// http filters
// ------------

func NewHttpRouterFilter() *envoyhttpcxnmgr.HttpFilter {
	filterConfig := envoyhttprouter.Router{}
	filterConfigPBStruct, err := envoyutil.MessageToStruct(&filterConfig)
	if err != nil {
		panic(fmt.Sprintf("error serializing http router filter: %v", err))
	}
	return &envoyhttpcxnmgr.HttpFilter{
		Name:   xdsconstants.HTTPRouterFilterName,
		Config: filterConfigPBStruct,
	}
}

// ---------------
// network filters
// ---------------

// HTTP connection manager

func NewRdsHttpConnectionManagerFilter(
	statPrefix string,
	routeConfigName string,
	httpFilters []*envoyhttpcxnmgr.HttpFilter) *envoylistener.Filter {
	filterConfig := envoyhttpcxnmgr.HttpConnectionManager{
		CodecType:  envoyhttpcxnmgr.AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &envoyhttpcxnmgr.HttpConnectionManager_Rds{
			Rds: &envoyhttpcxnmgr.Rds{
				ConfigSource: envoycore.ConfigSource{
					ConfigSourceSpecifier: &envoycore.ConfigSource_Ads{
						Ads: &envoycore.AggregatedConfigSource{},
					},
				},
				RouteConfigName: routeConfigName,
			},
		},
		HttpFilters: httpFilters,
	}
	filterConfigPBStruct, err := envoyutil.MessageToStruct(&filterConfig)
	if err != nil {
		panic(fmt.Sprintf("error serializing http connection manager filter: %v", err))
	}
	return &envoylistener.Filter{
		Name:   xdsconstants.HTTPConnectionManagerFilterName,
		Config: filterConfigPBStruct,
	}
}

func NewStaticHttpConnectionManagerFilter(
	statPrefix string,
	virtualHosts []envoyroute.VirtualHost,
	httpFilters []*envoyhttpcxnmgr.HttpFilter) *envoylistener.Filter {
	filterConfig := envoyhttpcxnmgr.HttpConnectionManager{
		CodecType:  envoyhttpcxnmgr.AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &envoyhttpcxnmgr.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoyv2.RouteConfiguration{
				VirtualHosts: virtualHosts,
			},
		},
		HttpFilters: httpFilters,
	}
	filterConfigPBStruct, err := envoyutil.MessageToStruct(&filterConfig)
	if err != nil {
		panic(fmt.Sprintf("error serializing http connection manager filter: %v", err))
	}
	return &envoylistener.Filter{
		Name:   xdsconstants.HTTPConnectionManagerFilterName,
		Config: filterConfigPBStruct,
	}
}

// TCP filter

func NewDeprecatedV1TCPProxyFilter(
	statPrefix string, routes *envoytcpproxy.TcpProxy_DeprecatedV1) *envoylistener.Filter {
	// https://godoc.org/github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2#TcpProxy

	// NOTE: inspiration drawn from:
	//       https://github.com/istio/istio/blob/master/pilot/pkg/networking/core/v1alpha3/networkfilter.go#L50

	deprecatedRoutes := &DeprecatedTCPRouteConfig{
		Routes: make([]*DeprecatedTCPRoute, 0, len(routes.Routes)),
	}

	for _, route := range routes.Routes {
		glog.V(4).Infof("route: %v", route)
		cidrList := make([]string, 0, len(route.DestinationIpList))
		// portList := make([]int32, 0, len(route.DestinationPorts))
		for _, cidr := range route.DestinationIpList {
			cidrList = append(cidrList, fmt.Sprintf("%s/%d", cidr.AddressPrefix, cidr.PrefixLen.Value))
		}
		// for _, port := range route.DestinationPorts {
		// 	portList = append(portList, port)
		// }
		// ports := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(portList)), ","), "[]")
		deprecatedRoutes.Routes = append(deprecatedRoutes.Routes, &DeprecatedTCPRoute{
			Cluster:           route.Cluster,
			DestinationIPList: cidrList,
			// DestinationPorts:  ports,
			DestinationPorts: route.DestinationPorts,
		})
	}

	filterConfig := &DeprecatedTCPProxyFilterConfig{
		StatPrefix:  statPrefix,
		RouteConfig: deprecatedRoutes,
	}

	filterConfigJSON, err := json.Marshal(filterConfig)
	if err != nil {
		panic(fmt.Sprintf("error trying to marshal V1 tcp proxy config to JSON: %v", err))
	}
	filterConfigStruct := &pbtypes.Struct{}
	if err := pbjson.Unmarshal(bytes.NewReader(filterConfigJSON), filterConfigStruct); err != nil {
		panic(fmt.Sprintf("error trying to unmarshal V1 tcp proxy JSON config to Struct: %v", err))
	}

	filterConfigValue := pbtypes.Value{
		Kind: &pbtypes.Value_StructValue{
			StructValue: filterConfigStruct,
		},
	}

	trueValue := pbtypes.Value{
		Kind: &pbtypes.Value_BoolValue{
			BoolValue: true,
		},
	}

	return &envoylistener.Filter{
		Name: xdsconstants.TCPProxyFilterName,
		Config: &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"deprecated_v1": &trueValue,
				"value":         &filterConfigValue,
			},
		},
	}
}

func NewTCPProxyFilter(statPrefix, clusterName string) *envoylistener.Filter {
	filterConfig := envoytcpproxy.TcpProxy{
		StatPrefix: statPrefix,
		Cluster:    clusterName,
	}

	filterConfigPBStruct, err := envoyutil.MessageToStruct(&filterConfig)
	if err != nil {
		panic(fmt.Sprintf("error serializing tcp proxy filter: %v", err))
	}

	return &envoylistener.Filter{
		Name:   xdsconstants.TCPProxyFilterName,
		Config: filterConfigPBStruct,
	}
}

// ----------------
// listener filters
// ----------------

func NewOriginalDestinationListenerFilter() *envoylistener.ListenerFilter {
	return &envoylistener.ListenerFilter{
		Name:   xdsconstants.OriginalDestinationListenerFilterName,
		Config: &pbtypes.Struct{},
	}
}
