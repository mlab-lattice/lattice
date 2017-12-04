package types

type Listener struct {
	Name    *string         `json:"name,omitempty"`
	Address string          `json:"address"`
	Filters []NetworkFilter `json:"filters"`
}

type NetworkFilter struct {
	Name   string       `json:"name"`
	Config FilterConfig `json:"config"`
}

type FilterConfig interface{}

type HTTPConnectionManagerConfig struct {
	CodecType   string       `json:"codec_type"`
	StatPrefix  string       `json:"stat_prefix"`
	RDS         *RDSConfig   `json:"rds,omitempty"`
	RouteConfig *RouteConfig `json:"route_config,omitempty"`
	Filters     []HTTPFilter `json:"filters"`
}

type RDSConfig struct {
	Cluster         string `json:"cluster"`
	RouteConfigName string `json:"route_config_name"`
	RefreshDelayMs  *int32 `json:"refresh_delay_ms,omitempty"`
}

type HTTPFilter struct {
	Name   string           `json:"name"`
	Config HTTPFilterConfig `json:"config"`
}

type HTTPFilterConfig interface{}

type RouterHTTPFilterConfig struct {
	DynamicStats *bool `json:"dynamic_stats,omitempty"`
}
