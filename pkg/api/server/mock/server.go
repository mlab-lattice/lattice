package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
)

const (
	mockServerAPIPort = 8876
)

func RunMockNewRestServer() {
	rest.RunNewRestServer(newMockBackend(), mockServerAPIPort, newMockSystemResolver(), "")
}
