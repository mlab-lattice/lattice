package v2

import (
	"github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/envoyproxy/go-control-plane/pkg/server"

	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Backend interface {
	log.Logger
	server.Callbacks
	envoycache.NodeHash

	Ready() bool
	Run(threadiness int) error
	XDSCache() envoycache.Cache
	SetXDSCacheSnapshot(id string, endpoints, clusters, routes, listeners []envoycache.Resource) error
	Services(serviceCluster string) (map[tree.NodePath]*Service, error)
}
