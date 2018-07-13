package constants

// TODO<GEB>: use constants in go-control-plane/util instead
const (
	// envoy network filter names
	HTTPConnectionManagerFilterName = "envoy.http_connection_manager"
	TCPProxyFilterName              = "envoy.tcp_proxy"

	// envoy listener filter names
	OriginalDestinationListenerFilterName = "envoy.listener.original_dst"

	// envoy http router filter name
	HTTPRouterFilterName = "envoy.router"

	// lattice egress listener names
	HTTPEgressListenerName = "egress-http"
	TCPEgressListenerName  = "egress-tcp"
)
