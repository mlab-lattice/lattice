package ads

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	// "github.com/gogo/protobuf/jsonpb"
	"github.com/golang/glog"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyendpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	// envoyhttprouter "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/router/v2"
	envoyhttpcxnmgr "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"
	envoyutil "github.com/envoyproxy/go-control-plane/pkg/util"
	// pbtypes "github.com/gogo/protobuf/types"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/constants"
	// xdstypes "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/types"
	xdsutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/util"
)

// XXX: rename to ServiceNode

type Service struct {
	Id string

	EnvoyNode *envoycore.Node

	lock sync.Mutex

	clusters  []envoycache.Resource
	endpoints []envoycache.Resource
	routes    []envoycache.Resource
	listeners []envoycache.Resource
}

func NewService(id string, envoyNode *envoycore.Node) *Service {
	return &Service{
		Id:        id,
		EnvoyNode: envoyNode,
	}
}

func (s *Service) Path() (tree.NodePath, error) {
	tnPath, err := tree.NodePathFromDomain(s.EnvoyNode.GetId())
	if err != nil {
		return "", err
	}
	return tnPath, nil
}

func (s *Service) Namespace() string {
	return s.EnvoyNode.GetCluster()
}

func (s *Service) getClusters(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	clusters := make([]envoycache.Resource, 0)

	for path, service := range systemServices {
		servicePath, err := s.Path()
		if err != nil {
			return nil, err
		}
		isLocalService := servicePath == path

		for componentName, component := range service.Components {
			for port := range component.Ports {
				clusterName := xdsutil.GetClusterNameForComponentPort(
					s.Namespace(), path, componentName, port)
				clusters = append(clusters, &envoyv2.Cluster{
					Name: clusterName,
					Type: envoyv2.Cluster_EDS,
					// TODO: figure out a good value for this
					ConnectTimeout: xdsconstants.ClusterConnectTimeout,
					LbPolicy:       envoyv2.Cluster_ROUND_ROBIN,
					EdsClusterConfig: &envoyv2.Cluster_EdsClusterConfig{
						EdsConfig: &envoycore.ConfigSource{
							ConfigSourceSpecifier: &envoycore.ConfigSource_Ads{
								Ads: &envoycore.AggregatedConfigSource{},
							},
						},
						ServiceName: clusterName,
					},
				})

				if isLocalService {
					clusterName = xdsutil.GetLocalClusterNameForComponentPort(
						s.Namespace(), path, componentName, port)
					clusters = append(clusters, &envoyv2.Cluster{
						Name: clusterName,
						Type: envoyv2.Cluster_STATIC,
						// TODO: figure out a good value for this
						ConnectTimeout: xdsconstants.ClusterConnectTimeout,
						LbPolicy:       envoyv2.Cluster_ROUND_ROBIN,
						Hosts: []*envoycore.Address{
							{
								Address: &envoycore.Address_SocketAddress{
									SocketAddress: &envoycore.SocketAddress{
										Protocol: envoycore.TCP,
										Address:  "127.0.0.1",
										PortSpecifier: &envoycore.SocketAddress_PortValue{
											PortValue: uint32(port),
										},
									},
								},
							},
						},
					})
				}
			}
		}
	}

	return clusters, nil
}

func (s *Service) getEndpoints(
	clusters []envoycache.Resource,
	systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
	endpoints := make([]envoycache.Resource, 0, len(clusters))
	for _, resource := range clusters {
		cluster := resource.(*envoyv2.Cluster)
		if cluster.EdsClusterConfig == nil {
			continue
		}
		_, path, componentName, port, err :=
			xdsutil.GetPartsFromClusterName(cluster.EdsClusterConfig.ServiceName)
		if err != nil {
			return nil, err
		}
		service, ok := systemServices[path]
		if !ok {
			return nil, fmt.Errorf("Invalid Service path <%v>", path)
		}
		component, ok := service.Components[componentName]
		if !ok {
			return nil, fmt.Errorf("Invalid Component name <%v>", componentName)
		}
		envoyPort, ok := component.Ports[port]
		if !ok {
			return nil, fmt.Errorf("Invalid Port <%v>", port)
		}
		addresses := make([]envoyendpoint.LbEndpoint, 0, len(service.IPAddresses))
		for _, address := range service.IPAddresses {
			addresses = append(addresses, envoyendpoint.LbEndpoint{
				Endpoint: &envoyendpoint.Endpoint{
					Address: &envoycore.Address{
						Address: &envoycore.Address_SocketAddress{
							SocketAddress: &envoycore.SocketAddress{
								Protocol: envoycore.TCP,
								Address:  address,
								PortSpecifier: &envoycore.SocketAddress_PortValue{
									PortValue: uint32(envoyPort),
								},
							},
						},
					},
				},
			})
		}
		endpoints = append(endpoints, &envoyv2.ClusterLoadAssignment{
			ClusterName: cluster.EdsClusterConfig.ServiceName,
			Endpoints: []envoyendpoint.LocalityLbEndpoints{
				{
					LbEndpoints: addresses,
				},
			},
		})
	}
	return endpoints, nil
}

func (s *Service) getListeners(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
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
													Cluster: xdsutil.GetLocalClusterNameForComponentPort(s.Namespace(), path, componentName, port),
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

func (s *Service) getRoutes(systemServices map[tree.NodePath]*xdsapi.Service) ([]envoycache.Resource, error) {
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
										Cluster: xdsutil.GetClusterNameForComponentPort(s.Namespace(), path, componentName, port),
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

func (s *Service) Update(backend xdsapi.Backend) error {
	glog.Info("Service.update called")
	// disallow concurrent updates to service state
	s.lock.Lock()
	defer s.lock.Unlock()

	systemServices, err := backend.SystemServices(s.Namespace())
	if err != nil {
		return err
	}

	clusters, err := s.getClusters(systemServices)
	if err != nil {
		return err
	}
	endpoints, err := s.getEndpoints(clusters, systemServices)
	if err != nil {
		return err
	}
	listeners, err := s.getListeners(systemServices)
	if err != nil {
		return err
	}
	routes, err := s.getRoutes(systemServices)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(clusters, s.clusters) ||
		!reflect.DeepEqual(endpoints, s.endpoints) ||
		!reflect.DeepEqual(listeners, s.listeners) ||
		!reflect.DeepEqual(routes, s.routes) {
		s.clusters = clusters
		s.endpoints = endpoints
		s.listeners = listeners
		s.routes = routes
		clustersJson, _ := json.MarshalIndent(s.clusters, "", "  ")
		endpointsJson, _ := json.MarshalIndent(s.endpoints, "", "  ")
		listenersJson, _ := json.MarshalIndent(s.listeners, "", "  ")
		routesJson, _ := json.MarshalIndent(s.routes, "", "  ")
		glog.Infof("Setting new snapshot for %v\nclusters\n%v\nendpoints\n%v\nlisteners\n%v\nroutes\n%v",
			s.Id, string(clustersJson), string(endpointsJson), string(listenersJson), string(routesJson))
		err := backend.SetXDSCacheSnapshot(s.Id, s.endpoints, s.clusters, s.routes, s.listeners)
		if err != nil {
			return err
		}
	}

	return nil
}
