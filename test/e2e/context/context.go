package context

import (
	"github.com/mlab-lattice/system/pkg/api/client/rest"
	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
)

type TestContextType struct {
	ClusterURL       string
	ClusterAPIClient clientv1.Interface
}

var TestContext TestContextType

func SetClusterURL(clusterURL string) {
	TestContext.ClusterURL = clusterURL
	TestContext.ClusterAPIClient = rest.NewClient(clusterURL)
}
