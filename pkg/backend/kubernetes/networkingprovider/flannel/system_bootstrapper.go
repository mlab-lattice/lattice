package flannel

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper/noop"
	"github.com/mlab-lattice/system/pkg/util/cli"
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

func ParseSystemBootstrapperFlags(vars []string) (*SystemBootstrapperOptions, error) {
	options := &SystemBootstrapperOptions{}
	flags := cli.EmbeddedFlag{
		Target: &options,
		Expected: map[string]cli.EmbeddedFlagValue{
			"cidr-block": {
				Required:     true,
				EncodingName: "CIDRBlock",
			},
		},
	}

	err := flags.Parse(vars)
	return options, err
}
