package rds

import (
	"fmt"

	xdsapi "github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/constants"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/types"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/util"
)

type RequestHandler struct {
	Backend xdsapi.Backend
}

type Response struct {
	VirtualHosts []types.VirtualHost `json:"virtual_hosts"`
}

func (r *RequestHandler) GetResponse(routeName, serviceCluster, serviceNode string) (*Response, error) {
	if routeName != constants.RouteNameEgress {
		return nil, fmt.Errorf("unexpected route name %v", routeName)
	}

	svcs, err := r.Backend.Services(serviceCluster)
	if err != nil {
		return nil, err
	}

	virtualHosts := []types.VirtualHost{}
	for path, svc := range svcs {
		for componentName, component := range svc.Components {
			for port := range component.Ports {
				pathDomain := path.ToDomain(true)
				domains := []string{fmt.Sprintf("%v:%v", pathDomain, port)}

				// Should be able to access an HTTP component on port 80 via either:
				//   - http://path.to.service:80
				//   - http://path.to.service
				if port == constants.PortHTTPDefault {
					domains = append(domains, pathDomain)
				}

				virtualHosts = append(virtualHosts, types.VirtualHost{
					Name:    string(path),
					Domains: domains,
					Routes: []types.VirtualHostRoute{
						{
							Prefix:  "/",
							Cluster: util.GetClusterNameForComponentPort(serviceCluster, path, componentName, port),
						},
					},
				})
			}
		}
	}

	resp := &Response{
		VirtualHosts: virtualHosts,
	}
	return resp, nil
}
