package base

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/util/terraform"
)

type Options struct {
	Config           latticev1.ConfigSpec
	MasterComponents MasterComponentOptions
	TerraformOptions TerraformOptions
}

type MasterComponentOptions struct {
	LatticeControllerManager LatticeControllerManagerOptions
	APIServer                APIServerOptions
}

type LatticeControllerManagerOptions struct {
	Image               string
	Args                []string
	TerraformModulePath string
}

type APIServerOptions struct {
	Image       string
	Port        int32
	HostNetwork bool
	Args        []string
}

type TerraformOptions struct {
	Backend terraform.BackendOptions
}

func NewBootstrapper(
	latticeID v1.LatticeID,
	cloudProviderName string,
	options *Options,
) (*DefaultBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	b := &DefaultBootstrapper{
		Options:           options,
		LatticeID:         latticeID,
		CloudProviderName: cloudProviderName,
	}
	return b, nil
}

type DefaultBootstrapper struct {
	Options           *Options
	LatticeID         v1.LatticeID
	CloudProviderName string
}

func (b *DefaultBootstrapper) BootstrapClusterResources(resources *bootstrapper.Resources) {
	b.namespaceResources(resources)
	b.crdResources(resources)
	b.configResources(resources)
	b.componentBuilderResources(resources)
	b.controllerManagerResources(resources)
	b.aPIServerResources(resources)
}
