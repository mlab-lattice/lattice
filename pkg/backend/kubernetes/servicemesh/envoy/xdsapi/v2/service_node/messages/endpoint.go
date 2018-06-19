package messages

import (
	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyendpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

func NewLbEndpoint(address *envoycore.Address) *envoyendpoint.LbEndpoint {
	return &envoyendpoint.LbEndpoint{
		Endpoint: &envoyendpoint.Endpoint{
			Address: address,
		},
	}
}

func NewClusterLoadAssignment(
	name string, lbEndpoints []envoyendpoint.LbEndpoint) *envoyv2.ClusterLoadAssignment {
	return &envoyv2.ClusterLoadAssignment{
		ClusterName: name,
		Endpoints: []envoyendpoint.LocalityLbEndpoints{
			{
				LbEndpoints: lbEndpoints,
			},
		},
	}
}
