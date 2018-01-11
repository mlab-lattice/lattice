package servicemesh

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"
)

type ClusterBootstrapperOptions struct {
	Envoy *envoy.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.Envoy != nil {
		return envoy.NewClusterBootstrapper(options.Envoy), nil
	}

	return nil, fmt.Errorf("must provide service mesh options")
}
