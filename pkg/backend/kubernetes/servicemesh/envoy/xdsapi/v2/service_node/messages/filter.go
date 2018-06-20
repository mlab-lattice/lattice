package messages

import (
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyhttprouter "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/router/v2"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoyutil "github.com/envoyproxy/go-control-plane/pkg/util"

	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
)

// filter chains

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

// http filters

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

// network filters

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
