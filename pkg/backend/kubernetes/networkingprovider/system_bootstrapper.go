package networkingprovider

import (
	"fmt"

	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/networkingprovider/flannel"
)

type SystemBootstrapperOptions struct {
	Flannel *flannel.SystemBootstrapperOptions
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) (systembootstrapper.Interface, error) {
	if options.Flannel != nil {
		return flannel.NewSystemBootstrapper(options.Flannel), nil
	}

	return nil, fmt.Errorf("must provide networking provider options")
}
