package context

import (
	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/apiserver/client/rest"
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
