package local

import (
	kubeclientset "k8s.io/client-go/kubernetes"
)

func NewBootstrapper(kubeClient kubeclientset.Interface) *DefaultBootstrapper {
	return &DefaultBootstrapper{
		KubeClient: kubeClient,
	}
}

type DefaultBootstrapper struct {
	KubeClient kubeclientset.Interface
}

func (b *DefaultBootstrapper) LocalBootstrap() error {
	bootstrapFuncs := []func() error{
		b.bootstrapLocalNode,
	}

	for _, bootstrapFunc := range bootstrapFuncs {
		if err := bootstrapFunc(); err != nil {
			return err
		}
	}
	return nil
}
