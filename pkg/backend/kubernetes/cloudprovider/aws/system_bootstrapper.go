package aws

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultAWSSystemBootstrapper {
	return &DefaultAWSSystemBootstrapper{
		DefaultBootstrapper: noop.NewBootstrapper(),
	}
}

type DefaultAWSSystemBootstrapper struct {
	*noop.DefaultBootstrapper
}

func SystemBootstrapperFlags() (command.Flags, *SystemBootstrapperOptions) {
	return nil, &SystemBootstrapperOptions{}
}

func ParseSystemBootstrapperFlags(vars []string) *SystemBootstrapperOptions {
	return &SystemBootstrapperOptions{}
}
