package local

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"

	"github.com/mlab-lattice/system/pkg/types"
)

type Options struct {
	DryRun           bool
	Config           crv1.ConfigSpec
	LocalComponents  LocalComponentOptions
}

type LocalComponentOptions struct {
	LocalDNSController 	LocalDNSControllerOptions
	LocalDNSServer		LocalDNSServerOptions
}

type LocalDNSControllerOptions struct {
	Image string
	Args  []string
}

type LocalDNSServerOptions struct {
	Image string
	Args  []string
}

func NewBootstrapper(
	ClusterID types.ClusterID,
	options *Options,
) (*DefaultBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	provider, err := crv1.GetProviderFromConfigSpec(&options.Config)
	if err != nil {
		return nil, err
	}

	b := &DefaultBootstrapper{
		Options:    options,
		Provider: 	provider,
		ClusterID:	ClusterID,
	}

	return b, nil
}

type DefaultBootstrapper struct {
	Options	 	*Options
	ClusterID	types.ClusterID
	Provider   	string
}

func (b *DefaultBootstrapper) BootstrapResources(resources *bootstrapper.Resources) {
	b.bootstrapLocalNode(resources)
	b.seedDNS(resources)
}