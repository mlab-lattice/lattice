package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
)

const (
	mockServerAPIPort = 8876
	mockServerAPIKey  = "abc"
)

func RunMockNewRestServer() {
	backend, err := newMockBackend()
	if err != nil {
		panic(err)
	}
	rest.RunNewRestServer(backend, newMockComponentResolver(), mockServerAPIPort, mockServerAPIKey)
}
