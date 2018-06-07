package types

type SocketAddress struct {
	Address   string `json:"address"`
	PortValue int32  `json:"port_value"`
}

// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/address.proto#envoy-api-msg-core-address
type Address struct {
	SocketAddress SocketAddress `json:"socket_address"`
}
