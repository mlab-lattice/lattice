package local

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	kubeclientset "k8s.io/client-go/kubernetes"
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
	kubeClient kubeclientset.Interface,
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
		KubeClient: kubeClient,
	}

	return b, nil
}

type DefaultBootstrapper struct {
	Options	 	*Options
	ClusterID	types.ClusterID
	Provider   	string
	KubeClient 	kubeclientset.Interface
}

func (b *DefaultBootstrapper) LocalBootstrap() ([]interface{}, error) {
	bootstrapFuncs := []func() ([]interface{}, error){
		b.bootstrapLocalNode,
		b.seedDNS,
	}

	var objects []interface{}
	for _, bootstrapFunc := range bootstrapFuncs {
		additionalObjects, err := bootstrapFunc()
		if err != nil {
			return nil, err
		}
		objects = append(objects, additionalObjects...)
	}
	return objects, nil
}
