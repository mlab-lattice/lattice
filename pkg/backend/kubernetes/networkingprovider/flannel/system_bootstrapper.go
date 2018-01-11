package flannel

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultFlannelSystemBootstrapper {
	return &DefaultFlannelSystemBootstrapper{
		DefaultBootstrapper: noop.NewBootstrapper(),
	}
}

type DefaultFlannelSystemBootstrapper struct {
	*noop.DefaultBootstrapper
}
