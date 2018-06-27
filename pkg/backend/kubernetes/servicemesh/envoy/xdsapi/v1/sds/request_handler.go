package sds

import (
	"fmt"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/types"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/util"
)

type RequestHandler struct {
	Backend xdsapi.Backend
}

type Response struct {
	Hosts []types.SDSHost `json:"hosts"`
}

func (r *RequestHandler) GetResponse(serviceName string) (*Response, error) {
	serviceCluster, path, componentName, port, err := util.GetPartsFromClusterName(serviceName)
	if err != nil {
		return nil, err
	}

	svcs, err := r.Backend.Services(serviceCluster)
	if err != nil {
		return nil, err
	}

	svc, ok := svcs[path]
	if !ok {
		return nil, fmt.Errorf("invalid Service path %v", path)
	}

	component, ok := svc.Containers[componentName]
	if !ok {
		return nil, fmt.Errorf("invalid Sidecar name %v", componentName)
	}

	envoyPort, ok := component.Ports[port]
	if !ok {
		return nil, fmt.Errorf("invalid Port %v", port)
	}

	hosts := []types.SDSHost{}
	for _, address := range svc.IPAddresses {
		hosts = append(hosts, types.SDSHost{
			IPAddress: address,
			Port:      envoyPort,
		})
	}

	resp := &Response{
		Hosts: hosts,
	}
	return resp, nil
}
