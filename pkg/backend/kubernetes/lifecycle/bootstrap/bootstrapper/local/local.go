package local

import (
	kubeclientset "k8s.io/client-go/kubernetes"
)

type Options struct {
	DryRun bool
}

func NewBootstrapper(options *Options, kubeClient kubeclientset.Interface) *DefaultBootstrapper {
	return &DefaultBootstrapper{
		Options:    options,
		KubeClient: kubeClient,
	}
}

type DefaultBootstrapper struct {
	Options *Options

	KubeClient kubeclientset.Interface
}

func (b *DefaultBootstrapper) LocalBootstrap() ([]interface{}, error) {
	bootstrapFuncs := []func() ([]interface{}, error){
		b.bootstrapLocalNode,
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
