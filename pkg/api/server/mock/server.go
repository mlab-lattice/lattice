package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
)

func RunMockNewRestServer() {
	rest.RunNewRestServer(newMockBackend(), 8876, newMockSystemResolver(), "")
}
