package v1

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
)

type Service struct {
	EgressPorts envoy.EnvoyEgressPorts
	Containers  map[string]Container
	IPAddresses []string
}

type Container struct {
	// Ports maps the Sidecar's ports to their envoy ports.
	Ports map[int32]int32
}
