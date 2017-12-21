package base

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/types"
)

type Options struct {
	DryRun           bool
	Config           crv1.ConfigSpec
	MasterComponents MasterComponentOptions
}

type MasterComponentOptions struct {
	LatticeControllerManager LatticeControllerManagerOptions
	ManagerAPI               ManagerAPIOptions
}

type LatticeControllerManagerOptions struct {
	Image string
	Args  []string
}

type ManagerAPIOptions struct {
	Image       string
	Port        int32
	HostNetwork bool
	Args        []string
}

func NewBootstrapper(
	clusterID types.ClusterID,
	cloudProviderName string,
	options *Options,
) (*DefaultBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	b := &DefaultBootstrapper{
		Options:           options,
		ClusterID:         clusterID,
		CloudProviderName: cloudProviderName,
	}
	return b, nil
}

type DefaultBootstrapper struct {
	Options           *Options
	ClusterID         types.ClusterID
	CloudProviderName string
}

func (b *DefaultBootstrapper) BootstrapResources(resources *bootstrapper.Resources) {
	b.namespaceResources(resources)
	b.crdResources(resources)
	b.configResources(resources)
	b.componentBuilderResources(resources)
	b.controllerManagerResources(resources)
	b.managerAPIResources(resources)
}
