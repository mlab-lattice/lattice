// This file was automatically generated by informer-gen

package v1

import (
	lattice_v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	versioned "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	internalinterfaces "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/listers/lattice/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// ServiceBuildInformer provides access to a shared informer and lister for
// ServiceBuilds.
type ServiceBuildInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ServiceBuildLister
}

type serviceBuildInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewServiceBuildInformer constructs a new informer for ServiceBuild type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewServiceBuildInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.LatticeV1().ServiceBuilds(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.LatticeV1().ServiceBuilds(namespace).Watch(options)
			},
		},
		&lattice_v1.ServiceBuild{},
		resyncPeriod,
		indexers,
	)
}

func defaultServiceBuildInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewServiceBuildInformer(client, meta_v1.NamespaceAll, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *serviceBuildInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&lattice_v1.ServiceBuild{}, defaultServiceBuildInformer)
}

func (f *serviceBuildInformer) Lister() v1.ServiceBuildLister {
	return v1.NewServiceBuildLister(f.Informer().GetIndexer())
}
