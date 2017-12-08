package cds

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	xdsapi "github.com/mlab-lattice/system/pkg/envoy/xdsapi/v1"
	"github.com/mlab-lattice/system/pkg/envoy/xdsapi/v1/constants"
	"github.com/mlab-lattice/system/pkg/envoy/xdsapi/v1/types"
	"github.com/mlab-lattice/system/pkg/envoy/xdsapi/v1/util"
)

type RequestHandler struct {
	Backend xdsapi.Backend
}

type Response struct {
	Clusters []types.Cluster `json:"clusters"`
}

func (r *RequestHandler) GetResponse(serviceCluster, serviceNode string) (*Response, error) {
	clusters := []types.Cluster{}
	svcs, err := r.Backend.Services()
	if err != nil {
		return nil, err
	}

	servicePath, err := tree.NodePathFromDomain(serviceNode)
	if err != nil {
		return nil, err
	}

	for path, svc := range svcs {
		isLocalService := servicePath == path

		for componentName, component := range svc.Components {
			for port := range component.Ports {
				clusterName := util.GetClusterNameForComponentPort(path, componentName, port)
				clusters = append(clusters, types.Cluster{
					Name: clusterName,
					Type: constants.ClusterTypeSDS,
					// TODO: figure out a good value for this
					ConnectTimeoutMs: 250,
					LBType:           constants.LBTypeRoundRobin,
					ServiceName:      clusterName,
				})

				if isLocalService {
					clusters = append(clusters, types.Cluster{
						Name: util.GetLocalClusterNameForComponentPort(path, componentName, port),
						Type: constants.ClusterTypeStatic,
						// TODO: figure out a good value for this
						ConnectTimeoutMs: 250,
						LBType:           constants.LBTypeRoundRobin,
						Hosts: []types.StaticHost{
							{
								URL: fmt.Sprintf("tcp://127.0.0.1:%v", port),
							},
						},
					})
				}
			}
		}
	}

	resp := &Response{
		Clusters: clusters,
	}
	return resp, nil
}
