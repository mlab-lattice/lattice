package messages

import (
	"time"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func NewEdsCluster(
	clusterName string,
	connectTimeout time.Duration,
	lbPolicy string) *envoyv2.Cluster {
	return &envoyv2.Cluster{
		Name: clusterName,
		Type: envoyv2.Cluster_EDS,
		// TODO: figure out a good value for this
		ConnectTimeout: connectTimeout,
		LbPolicy:       stringToClusterLbPolicy(lbPolicy),
		EdsClusterConfig: &envoyv2.Cluster_EdsClusterConfig{
			EdsConfig: &envoycore.ConfigSource{
				ConfigSourceSpecifier: &envoycore.ConfigSource_Ads{
					Ads: &envoycore.AggregatedConfigSource{},
				},
			},
			ServiceName: clusterName,
		},
	}
}

func NewStaticCluster(
	clusterName string,
	connectTimeout time.Duration,
	lbPolicy string,
	addresses []*envoycore.Address) *envoyv2.Cluster {
	return &envoyv2.Cluster{
		Name:           clusterName,
		Type:           envoyv2.Cluster_STATIC,
		ConnectTimeout: connectTimeout,
		LbPolicy:       stringToClusterLbPolicy(lbPolicy),
		Hosts:          addresses,
	}
}
