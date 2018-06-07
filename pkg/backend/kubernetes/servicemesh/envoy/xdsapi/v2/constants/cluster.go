package constants

const (
	// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cds.proto#envoy-api-enum-cluster-lbpolicy
	LBPolicyRoundRobin = "ROUND_ROBIN"

	// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cds.proto#envoy-api-enum-cluster-discoverytype
	ClusterTypeEDS    = "EDS"
	ClusterTypeStatic = "STATIC"

	ClusterConnectTimeout = "0.25s"
)
