package controllers

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	AddressController         = "address"
	BuildController           = "build"
	ContainerBuildController  = "containerbuild"
	JobController             = "job"
	NodePoolController        = "nodepool"
	ServiceController         = "service"
	SystemController          = "system"
	SystemLifecycleController = "systemlifecycle"
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
	NamespacePrefix string
	LatticeID       v1.LatticeID

	InternalDNSDomain string

	ComponentResolver resolver.Interface

	CloudProviderOptions *cloudprovider.Options
	ServiceMeshOptions   *servicemesh.Options

	// KubeInformerFactory gives access to base kubernetes kubeinformers.
	KubeInformerFactory kubeinformers.SharedInformerFactory

	// Need to create shared kubeinformers for each of our CRDs.
	LatticeInformerFactory latticeinformers.SharedInformerFactory

	KubeClientBuilder    KubeClientBuilder
	LatticeClientBuilder LatticeClientBuilder

	// Stop is the stop channel
	Stop <-chan struct{}
}
