package common

import (
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type ClientBuilder struct {
	Kubeconfig *rest.Config
}

func (cb ClientBuilder) ClientOrDie(name string) clientset.Interface {
	rest.AddUserAgent(cb.Kubeconfig, name)
	return clientset.NewForConfigOrDie(cb.Kubeconfig)
}

type Context struct {
	// InformerFactory gives access to base kubernetes informers.
	InformerFactory informers.SharedInformerFactory

	// Need to create shared informers for each of our CRDs.
	CRDInformers map[string]cache.SharedInformer

	LatticeResourceRestClient rest.Interface
	ClientBuilder             ClientBuilder

	// Stop is the stop channel
	Stop <-chan struct{}

	// Some controllers (cloud controllers) care about where
	// on the filesystem some terraform modules are.
	TerraformModulePath string
}

type Initializer func(context Context)
