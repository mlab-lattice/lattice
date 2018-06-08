package constants

import (
	"time"
)

const (
	// https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cds.proto#envoy-api-enum-cluster-discoverytype
	ClusterTypeEDS    = "EDS"
	ClusterTypeStatic = "STATIC"

	ClusterConnectTimeout = time.Duration(250) * time.Millisecond
)
