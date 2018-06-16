package constants

import (
	"time"
)

const (
	ClusterConnectTimeout     = time.Duration(250) * time.Millisecond
	ClusterLbPolicyRoundRobin = "ROUND_ROBIN"
)
