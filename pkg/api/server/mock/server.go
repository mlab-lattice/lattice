package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/mock/backend"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
)

func RunMockNewRestServer(port int32, apiAuthKey, workDirectory string) {
	b := backend.NewMockBackend()

	componentResolver, err := resolver.NewComponentResolver(
		workDirectory,
		true,
		resolver.NewMemoryTemplateStore(),
		resolver.NewMemorySecretStore(),
	)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(b, componentResolver, port, apiAuthKey)
}
