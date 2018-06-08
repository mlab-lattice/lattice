package service_node

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"
	envoyutil "github.com/envoyproxy/go-control-plane/pkg/util"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
)

func (s *ServiceNode) getListeners(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	var err error

	listeners := make([]envoycache.Resource, 0)

	path, err := s.Path()
	if err != nil {
		return nil, err
	}

	service, ok := systemServices[path]
	if !ok {
		return nil, fmt.Errorf("Invalid Service path <%v>", path)
	}

	// var configBytes []byte

	// httpFilterConfig := envoyhttprouter.Router{}
	// httpFilterConfigPBStruct, err := envoyutil.MessageToStruct(&httpFilterConfig)
	// if err != nil {
	// 	return nil, err
	// }

	// defined as protobuf type.Struct in its parent protobuf
	filterConfig := envoyhttpcxnmgr.HttpConnectionManager{
		CodecType:  envoyhttpcxnmgr.AUTO,
		StatPrefix: "egress",
		RouteSpecifier: &envoyhttpcxnmgr.HttpConnectionManager_Rds{
			Rds: &envoyhttpcxnmgr.Rds{
				ConfigSource: envoycore.ConfigSource{
					ConfigSourceSpecifier: &envoycore.ConfigSource_Ads{
						Ads: &envoycore.AggregatedConfigSource{},
					},
				},
				RouteConfigName: xdsconstants.RouteNameEgress,
			},
		},
		HttpFilters: []*envoyhttpcxnmgr.HttpFilter{
			{
				Name: xdsconstants.HTTPFilterRouterName,
				// type.Struct
				// Config: httpFilterConfigPBStruct,
			},
		},
	}
	filterConfigPBStruct, err := envoyutil.MessageToStruct(&filterConfig)
	if err != nil {
		return nil, err
	}

	listeners = append(listeners, &envoyv2.Listener{
		Name: "egress",
		Address: envoycore.Address{
			Address: &envoycore.Address_SocketAddress{
				SocketAddress: &envoycore.SocketAddress{
					Protocol:      envoycore.TCP,
					Address:       "0.0.0.0",
					PortSpecifier: &envoycore.SocketAddress_PortValue{PortValue: uint32(service.EgressPort)},
				},
			},
		},
		FilterChains: []envoylistener.FilterChain{
			{
				Filters: []envoylistener.Filter{
					{
						Name: xdsconstants.FilterHTTPConnectionManagerName,
						// type.Struct
						Config: filterConfigPBStruct,
					},
				},
			},
		},
	})

	// There's a listener for each port of Service, listening on the port's EnvoyPort
	for componentName, component := range service.Components {
		for port, envoyPort := range component.Ports {
			listenerName := fmt.Sprintf("%v %v port %v ingress", path, componentName, port)
			filterConfig := envoyhttpcxnmgr.HttpConnectionManager{
				CodecType:  envoyhttpcxnmgr.AUTO,
				StatPrefix: listenerName,
				RouteSpecifier: &envoyhttpcxnmgr.HttpConnectionManager_RouteConfig{
					RouteConfig: &envoyv2.RouteConfiguration{
						VirtualHosts: []envoyroute.VirtualHost{
							{
								Name:    fmt.Sprintf("%v %v port %v", path, componentName, port),
								Domains: []string{"*"},
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
													Cluster: xdsutil.GetLocalClusterNameForComponentPort(
														s.ServiceCluster(), path, componentName, port),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				HttpFilters: []*envoyhttpcxnmgr.HttpFilter{
					{
						Name: xdsconstants.HTTPFilterRouterName,
						// Config: httpFilterConfigPBStruct,
					},
				},
				// FIXME: add health_check filter
				// FIXME: look into other filters (buffer, potentially add fault injection for testing)
			}
			filterConfigPBStruct, err = envoyutil.MessageToStruct(&filterConfig)
			if err != nil {
				return nil, err
			}

			listeners = append(listeners, &envoyv2.Listener{
				Name: listenerName,
				Address: envoycore.Address{
					Address: &envoycore.Address_SocketAddress{
						SocketAddress: &envoycore.SocketAddress{
							Protocol:      envoycore.TCP,
							Address:       "0.0.0.0",
							PortSpecifier: &envoycore.SocketAddress_PortValue{PortValue: uint32(envoyPort)},
						},
					},
				},
				FilterChains: []envoylistener.FilterChain{
					{
						Filters: []envoylistener.Filter{
							{
								Name:   xdsconstants.FilterHTTPConnectionManagerName,
								Config: filterConfigPBStruct,
							},
						},
					},
				},
			})
		}
	}
	return listeners, nil
}
