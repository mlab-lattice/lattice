package messages

import (
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func NewTcpSocketAddress(address string, port int32) *envoycore.Address {
	return &envoycore.Address{
		Address: &envoycore.Address_SocketAddress{
			SocketAddress: &envoycore.SocketAddress{
				Protocol: envoycore.TCP,
				Address:  address,
				PortSpecifier: &envoycore.SocketAddress_PortValue{
					PortValue: uint32(port),
				},
			},
		},
	}
}
