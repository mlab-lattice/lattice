package cloud

import (
	"fmt"

	kubeclientset "k8s.io/client-go/kubernetes"
)

type Options struct {
	DryRun     bool
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

func (b *DefaultBootstrapper) CloudBootstrap() ([]interface{}, error) {
	bootstrapFuncs := []func() ([]interface{}, error){
		b.bootstrapNetworking,
	}

	objects := []interface{}{}
	for _, bootstrapFunc := range bootstrapFuncs {
		additionalObjects, err := bootstrapFunc()
		if err != nil {
			return nil, err
		}
		objects = append(objects, additionalObjects...)
	}
	return objects, nil
}
