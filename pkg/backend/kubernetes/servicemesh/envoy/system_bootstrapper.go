package envoy

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultEnvoySystemBootstrapper {
	return &DefaultEnvoySystemBootstrapper{
		DefaultBootstrapper: noop.NewBootstrapper(),
	}
}

type DefaultEnvoySystemBootstrapper struct {
	*noop.DefaultBootstrapper
}

func SystemBootstrapperFlags() (cli.Flags, *SystemBootstrapperOptions) {
	return nil, &SystemBootstrapperOptions{}
}

func ParseSystemBootstrapperFlags(vars []string) *SystemBootstrapperOptions {
	return &SystemBootstrapperOptions{}
}
