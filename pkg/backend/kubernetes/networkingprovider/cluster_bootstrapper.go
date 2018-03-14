package networkingprovider

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/none"
)

type ClusterBootstrapperOptions struct {
	Flannel *flannel.ClusterBootstrapperOptions
	None    *none.ClusterBootstrapperOptions
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.Flannel != nil {
		return flannel.NewClusterBootstrapper(options.Flannel), nil
	}

	if options.None != nil {
		return none.NewClusterBootstrapper(options.None), nil
	}

	return nil, fmt.Errorf("must provide networking provider options")
}
