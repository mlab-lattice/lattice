package networkingprovider

import (
	"fmt"

	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
)

const (
	Flannel = "flannel"
)

type Interface interface {
	clusterbootstrapper.Interface
}

type Options struct {
	Flannel *flannel.Options
}

func NewNetworkingProvider(options *Options) (Interface, error) {
	if options.Flannel != nil {
		return flannel.NewFlannelNetworkingProvider(options.Flannel), nil
	}

	return nil, fmt.Errorf("no networking provider configuration set")
}
