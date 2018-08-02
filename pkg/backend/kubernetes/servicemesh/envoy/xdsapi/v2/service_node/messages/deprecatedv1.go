package messages

// NOTE: the following were taken from:
//       https://github.com/istio/istio/blob/master/pilot/pkg/networking/core/v1alpha3/networkfilter.go#L50

// DeprecatedTCPRoute definition
type DeprecatedTCPRoute struct {
	Cluster           string   `json:"cluster"`
	DestinationIPList []string `json:"destination_ip_list,omitempty"`
	DestinationPorts  string   `json:"destination_ports,omitempty"`
	SourceIPList      []string `json:"source_ip_list,omitempty"`
	SourcePorts       string   `json:"source_ports,omitempty"`
}

// DeprecatedTCPRouteConfig (or generalize as RouteConfig or L4RouteConfig for TCP/UDP?)
type DeprecatedTCPRouteConfig struct {
	Routes []*DeprecatedTCPRoute `json:"routes"`
}

// DeprecatedTCPProxyFilterConfig definition
type DeprecatedTCPProxyFilterConfig struct {
	StatPrefix  string                    `json:"stat_prefix"`
	RouteConfig *DeprecatedTCPRouteConfig `json:"route_config"`
}
