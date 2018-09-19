package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
)

func RunMockNewRestServer(port int32, options *rest.ServerOptions) {
	backend, err := newMockBackend()
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(backend, newMockComponentResolver(), port, options)
}
