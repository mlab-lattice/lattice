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

type HttpConnectionManagerConfig struct {
	CodecType   string       `json:"codec_type"`
	StatPrefix  string       `json:"stat_prefix"`
	RDS         *RDSConfig   `json:"rds,omitempty"`
	RouteConfig *RouteConfig `json:"route_config,omitempty"`
	Filters     []HttpFilter `json:"filters"`
}

type RDSConfig struct {
	Cluster         string `json:"cluster"`
	RouteConfigName string `json:"route_config_name"`
	RefreshDelayMs  *int32 `json:"refresh_delay_ms,omitempty"`
}

type HttpFilter struct {
	Name   string           `json:"name"`
	Config HttpFilterConfig `json:"config"`
}

type HttpFilterConfig interface{}

type RouterHttpFilterConfig struct {
	DynamicStats *bool `json:"dynamic_stats,omitempty"`
}
