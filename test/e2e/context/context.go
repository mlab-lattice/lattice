package context

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

type TestContextType struct {
	ClusterURL       string
	ClusterAPIClient client.Interface
}

var TestContext TestContextType

func SetClusterURL(clusterURL string) {
	TestContext.ClusterURL = clusterURL
	TestContext.ClusterAPIClient = rest.NewClient(clusterURL)
}
