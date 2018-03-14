package flannel

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultSystemBootstrapper {
	return &DefaultSystemBootstrapper{
		DefaultBootstrapper: noop.NewBootstrapper(),
	}
}

type DefaultSystemBootstrapper struct {
	*noop.DefaultBootstrapper
}

func ParseSystemBootstrapperFlags(vars []string) (*SystemBootstrapperOptions, error) {
	return &SystemBootstrapperOptions{}, nil
}
