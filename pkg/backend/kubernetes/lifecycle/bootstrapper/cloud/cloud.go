package cloud

import (
	"fmt"

	kubeclientset "k8s.io/client-go/kubernetes"
)

type Options struct {
	Networking *NetworkingOptions
}

type NetworkingOptions struct {
	Flannel *FlannelOptions
}

type FlannelOptions struct {
	NetworkCIDRBlock string
}

func NewBootstrapper(options *Options, kubeClient kubeclientset.Interface) (*DefaultBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	b := &DefaultBootstrapper{
		Options:    options,
		KubeClient: kubeClient,
	}
	return b, nil
}

type DefaultBootstrapper struct {
	Options *Options

	KubeClient kubeclientset.Interface
}

func (b *DefaultBootstrapper) CloudBootstrap() error {
	bootstrapFuncs := []func() error{
		b.bootstrapNetworking,
	}

	for _, bootstrapFunc := range bootstrapFuncs {
		if err := bootstrapFunc(); err != nil {
			return err
		}
	}
	return nil
}
