package cds

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
	Clusters []types.Cluster `json:"clusters"`
}

func (r *RequestHandler) GetResponse(serviceCluster, serviceNode string) (*Response, error) {
	services, err := r.Backend.Services(serviceCluster)
	if err != nil {
		return nil, err
	}

	servicePath, err := tree.NewPathFromDomain(serviceNode)
	if err != nil {
		return nil, err
	}

	var clusters []types.Cluster
	for path, svc := range services {
		isLocalService := servicePath == path

		for componentName, component := range svc.Containers {
			for port := range component.Ports {
				clusterName := util.GetClusterNameForComponentPort(serviceCluster, path, componentName, port)
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
						Name: util.GetLocalClusterNameForComponentPort(serviceCluster, path, componentName, port),
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
