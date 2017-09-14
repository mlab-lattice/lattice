package common

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/provider"

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
	Provider provider.Interface

	// InformerFactory gives access to base kubernetes informers.
	InformerFactory informers.SharedInformerFactory

	// Need to create shared informers for each of our CRDs.
	CRDInformers map[string]cache.SharedInformer

	LatticeResourceRestClient rest.Interface
	ClientBuilder             ClientBuilder

	// Stop is the stop channel
	Stop <-chan struct{}
}

type Initializer func(context Context)
