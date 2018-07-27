package servicemesh

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
)

const (
	Envoy           = envoy.Envoy
	AnnotationKeyIP = "envoy.servicemesh.lattice.mlab.com/ip"
)
