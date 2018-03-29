package context

import (
	"github.com/mlab-lattice/system/pkg/api/client"
	"github.com/mlab-lattice/system/pkg/api/client/rest"
)

type TestContextType struct {
	LatticeURL       string
	LatticeAPIClient client.Interface
}

var TestContext TestContextType

func SetClusterURL(clusterURL string) {
	TestContext.LatticeURL = clusterURL
	TestContext.LatticeAPIClient = rest.NewClient(clusterURL)
}
