package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
)

const (
	mockServerAPIPort = 8876
	mockServerAPIKey  = "abc"
)

func RunMockNewRestServer() {
	rest.RunNewRestServer(newMockBackend(), mockServerAPIPort, newMockSystemResolver(), mockServerAPIKey)
}
