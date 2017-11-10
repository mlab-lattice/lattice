package sds

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/envoy"
	"github.com/mlab-lattice/system/pkg/envoy/xds-api/types"
	"github.com/mlab-lattice/system/pkg/envoy/xds-api/util"
)

type RequestHandler struct {
	Backend envoy.Backend
}

type Response struct {
	Hosts []types.SDSHost `json:"hosts"`
}

func (r *RequestHandler) GetResponse(serviceName string) (*Response, error) {
	path, componentName, port, err := util.GetPartsFromClusterName(serviceName)
	if err != nil {
		return nil, err
	}

	svcs, err := r.Backend.Services()
	if err != nil {
		return nil, err
	}

	svc, ok := svcs[path]
	if !ok {
		return nil, fmt.Errorf("invalid Service path %v", path)
	}

	component, ok := svc.Components[componentName]
	if !ok {
		return nil, fmt.Errorf("invalid Component name %v", componentName)
	}

	envoyPort, ok := component.Ports[port]
	if !ok {
		return nil, fmt.Errorf("invalid Port %v", port)
	}

	hosts := []types.SDSHost{}
	for _, address := range svc.IPAddresses {
		hosts = append(hosts, types.SDSHost{
			IpAddress: address,
			Port:      envoyPort,
		})
	}

	resp := &Response{
		Hosts: hosts,
	}
	return resp, nil
}
