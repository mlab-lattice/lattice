// This file was automatically generated by lister-gen

package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ServiceBuildLister helps list ServiceBuilds.
type ServiceBuildLister interface {
	// List lists all ServiceBuilds in the indexer.
	List(selector labels.Selector) (ret []*v1.ServiceBuild, err error)
	// ServiceBuilds returns an object that can list and get ServiceBuilds.
	ServiceBuilds(namespace string) ServiceBuildNamespaceLister
	ServiceBuildListerExpansion
}

// serviceBuildLister implements the ServiceBuildLister interface.
type serviceBuildLister struct {
	indexer cache.Indexer
}

// NewServiceBuildLister returns a new ServiceBuildLister.
func NewServiceBuildLister(indexer cache.Indexer) ServiceBuildLister {
	return &serviceBuildLister{indexer: indexer}
}

// List lists all ServiceBuilds in the indexer.
func (s *serviceBuildLister) List(selector labels.Selector) (ret []*v1.ServiceBuild, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceBuild))
	})
	return ret, err
}

// ServiceBuilds returns an object that can list and get ServiceBuilds.
func (s *serviceBuildLister) ServiceBuilds(namespace string) ServiceBuildNamespaceLister {
	return serviceBuildNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceBuildNamespaceLister helps list and get ServiceBuilds.
type ServiceBuildNamespaceLister interface {
	// List lists all ServiceBuilds in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.ServiceBuild, err error)
	// Get retrieves the ServiceBuild from the indexer for a given namespace and name.
	Get(name string) (*v1.ServiceBuild, error)
	ServiceBuildNamespaceListerExpansion
}

// serviceBuildNamespaceLister implements the ServiceBuildNamespaceLister
// interface.
type serviceBuildNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceBuilds in the indexer for a given namespace.
func (s serviceBuildNamespaceLister) List(selector labels.Selector) (ret []*v1.ServiceBuild, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceBuild))
	})
	return ret, err
}

// Get retrieves the ServiceBuild from the indexer for a given namespace and name.
func (s serviceBuildNamespaceLister) Get(name string) (*v1.ServiceBuild, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("servicebuild"), name)
	}
	return obj.(*v1.ServiceBuild), nil
}
