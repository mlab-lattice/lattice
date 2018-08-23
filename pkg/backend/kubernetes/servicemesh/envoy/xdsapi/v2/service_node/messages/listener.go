package messages

import (
	pbtypes "github.com/gogo/protobuf/types"

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

func NewOriginalDestinationListener(
	name string,
	address *envoycore.Address,
	filterChains []envoylistener.FilterChain) *envoyv2.Listener {
	listener := NewListener(name, address, filterChains)
	// deprecated V1 way to enable original destination filtering
	listener.UseOriginalDst = &pbtypes.BoolValue{Value: true}
	// V2 filter chain way to enable original destination filtering (not currently used)
	listener.ListenerFilters = []envoylistener.ListenerFilter{
		*NewOriginalDestinationListenerFilter(),
	}
	return listener
}
