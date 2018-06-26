package constants

import (
	"time"
)

const (
	ClusterConnectTimeout     = time.Duration(250) * time.Millisecond
	ClusterLBPolicyRoundRobin = "ROUND_ROBIN"
)
