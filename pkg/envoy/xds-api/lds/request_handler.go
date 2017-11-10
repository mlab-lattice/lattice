package lds

import (
	"fmt"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	"github.com/mlab-lattice/system/pkg/envoy"
	"github.com/mlab-lattice/system/pkg/envoy/xds-api/constants"
	"github.com/mlab-lattice/system/pkg/envoy/xds-api/types"
	"github.com/mlab-lattice/system/pkg/envoy/xds-api/util"
)

type RequestHandler struct {
	Backend envoy.Backend
}

type Response struct {
	Listeners []types.Listener `json:"listeners"`
}

func (r *RequestHandler) GetResponse(serviceCluster, serviceNode string) (*Response, error) {
	path, err := systemtree.NodePathFromDomain(serviceNode)
	if err != nil {
		return nil, err
	}

	svcs, err := r.Backend.Services()
	if err != nil {
		return nil, err
	}

	service, ok := svcs[path]
	if !ok {
		return nil, fmt.Errorf("invalid Service path %v", path)
	}

	// There's a single egress listener listening for traffic from within the pod.
	egress := "egress"
	listeners := []types.Listener{
		{
			Name:    &egress,
			Address: fmt.Sprintf("tcp://0.0.0.0:%v", service.EgressPort),
			Filters: []types.NetworkFilter{
				{
					Name: constants.FilterNameHttpConnectionManager,
					Config: types.HttpConnectionManagerConfig{
						CodecType:  constants.CodecTypeAuto,
						StatPrefix: egress,
						RDS: &types.RDSConfig{
							Cluster:         constants.XdsApiClusterName,
							RouteConfigName: constants.RouteNameEgress,
						},
						Filters: []types.HttpFilter{
							{
								Name:   constants.HttpFilterNameRouter,
								Config: types.RouterHttpFilterConfig{},
							},
						},
					},
				},
			},
		},
	}

	// There's a listener for each port of Service, listening on the port's EnvoyPort
	for componentName, component := range service.Components {
		for port, envoyPort := range component.Ports {
			listenerName := fmt.Sprintf("%v %v port %v ingress", path, componentName, port)
			listeners = append(listeners, types.Listener{
				Name:    &listenerName,
				Address: fmt.Sprintf("tcp://0.0.0.0:%v", envoyPort),
				Filters: []types.NetworkFilter{
					{
						Name: constants.FilterNameHttpConnectionManager,
						Config: types.HttpConnectionManagerConfig{
							CodecType:  constants.CodecTypeAuto,
							StatPrefix: listenerName,
							RouteConfig: &types.RouteConfig{
								VirtualHosts: []types.VirtualHost{
									{
										Name:    fmt.Sprintf("%v %v port %v", path, componentName, port),
										Domains: []string{"*"},
										Routes: []types.VirtualHostRoute{
											{
												Prefix:  "/",
												Cluster: util.GetLocalClusterNameForComponentPort(path, componentName, port),
											},
										},
									},
								},
							},
							Filters: []types.HttpFilter{
								{
									Name:   constants.HttpFilterNameRouter,
									Config: types.RouterHttpFilterConfig{},
								},
								// FIXME: add health_check filter
								// FIXME: look into other filters (buffer, potentially add fault injection for testing)
							},
						},
					},
				},
			})
		}
	}

	resp := &Response{
		Listeners: listeners,
	}
	return resp, nil
}
