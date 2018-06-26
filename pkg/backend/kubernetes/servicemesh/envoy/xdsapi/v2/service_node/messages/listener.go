package messages

import (
	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
)

func NewListener(
	name string,
	address *envoycore.Address,
	filterChains []envoylistener.FilterChain) *envoyv2.Listener {
	return &envoyv2.Listener{
		Name:         name,
		Address:      *address,
		FilterChains: filterChains,
	}
}
