package lds

import (
	"fmt"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/constants"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/types"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/util"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type RequestHandler struct {
	Backend xdsapi.Backend
}

type Response struct {
	Listeners []types.Listener `json:"listeners"`
}

func (r *RequestHandler) GetResponse(serviceCluster, serviceNode string) (*Response, error) {
	path, err := tree.NewNodePathFromDomain(serviceNode)
	if err != nil {
		return nil, err
	}

	svcs, err := r.Backend.Services(serviceCluster)
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
			Address: fmt.Sprintf("tcp://0.0.0.0:%v", service.EgressPorts.HTTP),
			Filters: []types.NetworkFilter{
				{
					Name: constants.FilterNameHTTPConnectionManager,
					Config: types.HTTPConnectionManagerConfig{
						CodecType:  constants.CodecTypeAuto,
						StatPrefix: egress,
						RDS: &types.RDSConfig{
							Cluster:         constants.HTTPXDSApiClusterName,
							RouteConfigName: constants.RouteNameEgress,
						},
						Filters: []types.HTTPFilter{
							{
								Name:   constants.HTTPFilterNameRouter,
								Config: types.RouterHTTPFilterConfig{},
							},
						},
					},
				},
			},
		},
	}

	// There's a listener for each port of Service, listening on the port's EnvoyPort
	for componentName, component := range service.Containers {
		for port, envoyPort := range component.Ports {
			listenerName := fmt.Sprintf("%v %v port %v ingress", path, componentName, port)
			listeners = append(listeners, types.Listener{
				Name:    &listenerName,
				Address: fmt.Sprintf("tcp://0.0.0.0:%v", envoyPort),
				Filters: []types.NetworkFilter{
					{
						Name: constants.FilterNameHTTPConnectionManager,
						Config: types.HTTPConnectionManagerConfig{
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
												Cluster: util.GetLocalClusterNameForComponentPort(serviceCluster, path, componentName, port),
											},
										},
									},
								},
							},
							Filters: []types.HTTPFilter{
								{
									Name:   constants.HTTPFilterNameRouter,
									Config: types.RouterHTTPFilterConfig{},
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
