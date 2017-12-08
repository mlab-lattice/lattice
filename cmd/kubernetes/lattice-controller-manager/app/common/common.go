package common

import (
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/informers/externalversions"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

type KubeClientBuilder struct {
	Kubeconfig *rest.Config
}

func (cb KubeClientBuilder) ClientOrDie(name string) kubeclientset.Interface {
	rest.AddUserAgent(cb.Kubeconfig, name)
	return kubeclientset.NewForConfigOrDie(cb.Kubeconfig)
}

type LatticeClientBuilder struct {
	Kubeconfig *rest.Config
}

func (cb LatticeClientBuilder) ClientOrDie(name string) latticeclientset.Interface {
	rest.AddUserAgent(cb.Kubeconfig, name)
	return latticeclientset.NewForConfigOrDie(cb.Kubeconfig)
}

type Context struct {
	// KubeInformerFactory gives access to base kubernetes kubeinformers.
	KubeInformerFactory kubeinformers.SharedInformerFactory

	// Need to create shared kubeinformers for each of our CRDs.
	LatticeInformerFactory latticeinformers.SharedInformerFactory

	KubeClientBuilder    KubeClientBuilder
	LatticeClientBuilder LatticeClientBuilder

	// Stop is the stop channel
	Stop <-chan struct{}

	// Some controllers (cloud controllers) care about where
	// on the filesystem some terraform modules are.
	TerraformModulePath string
}

type CRInformers struct {
	ComponentBuild cache.SharedInformer
	Config         cache.SharedInformer
	Service        cache.SharedInformer
	ServiceBuild   cache.SharedInformer
	System         cache.SharedInformer
	SystemBuild    cache.SharedInformer
	SystemRollout  cache.SharedInformer
	SystemTeardown cache.SharedInformer
}

func (cri *CRInformers) Start(stopCh <-chan struct{}) {
	crInformers := []struct {
		name     string
		informer *cache.SharedInformer
	}{
		{
			name:     "component-build",
			informer: &cri.ComponentBuild,
		},
		{
			name:     "config",
			informer: &cri.Config,
		},
		{
			name:     "service",
			informer: &cri.Service,
		},
		{
			name:     "service-build",
			informer: &cri.ServiceBuild,
		},
		{
			name:     "system",
			informer: &cri.System,
		},
		{
			name:     "system-build",
			informer: &cri.SystemBuild,
		},
		{
			name:     "system-rollout",
			informer: &cri.SystemRollout,
		},
		{
			name:     "system-teardown",
			informer: &cri.SystemTeardown,
		},
	}

	for _, informer := range crInformers {
		glog.V(4).Infof("Starting %v informer", informer.name)
		go (*informer.informer).Run(stopCh)
	}
}

type Initializer func(context Context)
