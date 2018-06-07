package types

// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/config_source.proto#envoy-api-msg-core-configsource
type EDSConfig struct {
	ADS struct{} `json:"ads"`
}

// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cds.proto#envoy-api-msg-cluster-edsclusterconfig
type EDSClusterConfig struct {
	EDSConfig   EDSConfig `json:"eds_config"`
	ServiceName string    `json:"service_name",omitempty`
}

// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cds.proto#cluster
type Cluster struct {
	Name             string           `json:"name"`
	Type             string           `json:"type"`
	ConnectTimeout   string           `json:"connect_timeout"`
	LBPolicy         string           `json:"lb_policy"`
	Hosts            []Address        `json:"hosts,omitempty"`
	EDSClusterConfig EDSClusterConfig `json:"eds_cluster_config",omitempty`
	// TODO: reexamine other fields
}
