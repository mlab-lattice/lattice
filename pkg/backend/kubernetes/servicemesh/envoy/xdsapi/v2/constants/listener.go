package constants

// XXX: use constants in go-control-plane/util instead
const (
	HTTPConnectionManagerFilterName = "envoy.http_connection_manager"
	TCPConnectionManagerFilterName  = "envoy.tcp_proxy"

	HTTPRouterFilterName = "envoy.router"

	HTTPEgressListenerName = "egress-http"
	TCPEgressListenerName  = "egress-tcp"
)
