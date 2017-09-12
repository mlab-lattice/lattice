package app

import (
	"time"

	latticeresource "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

func Run(kubeconfig, provider string) {
	// TODO: create in-cluster config if in cluster
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		panic(err)
	}

	// TODO: setting stop as nil for now, won't actually need it until leader-election is used
	ctx := CreateControllerContext(rest.Interface(latticeResourceClient), config, provider, nil)
	glog.V(1).Info("Starting controllers")
	StartControllers(ctx, CreateControllerInitializers())

	glog.V(4).Info("Starting informer factory informers")
	ctx.InformerFactory.Start(ctx.Stop)

	for resource, informer := range ctx.CRDInformers {
		glog.V(4).Infof("Starting %v informer", resource)
		go informer.Run(ctx.Stop)

	}

	select {}
}

type ClientBuilder struct {
	kubeconfig *rest.Config
}

func (cb ClientBuilder) ClientOrDie(name string) clientset.Interface {
	rest.AddUserAgent(cb.kubeconfig, name)
	return clientset.NewForConfigOrDie(cb.kubeconfig)
}

type ControllerContext struct {
	Provider string

	// InformerFactory gives access to base kubernetes informers.
	InformerFactory informers.SharedInformerFactory

	// Need to create shared informers for each of our CRDs.
	CRDInformers map[string]cache.SharedInformer

	LatticeResourceRestClient rest.Interface
	ClientBuilder             ClientBuilder

	// Stop is the stop channel
	Stop <-chan struct{}
}

func CreateControllerContext(
	latticeResourceClient rest.Interface,
	kubeconfig *rest.Config,
	provider string,
	stop <-chan struct{},
) ControllerContext {
	cb := ClientBuilder{
		kubeconfig: kubeconfig,
	}

	versionedClient := cb.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactory(versionedClient, time.Duration(12*time.Hour))

	return ControllerContext{
		Provider:                  provider,
		InformerFactory:           sharedInformers,
		CRDInformers:              getCRDInformers(latticeResourceClient),
		LatticeResourceRestClient: latticeResourceClient,
		ClientBuilder:             cb,

		Stop: stop,
	}
}

func getCRDInformers(latticeResourceClient rest.Interface) map[string]cache.SharedInformer {
	// FIXME: defaultResync blindly taken from k8s.io/kubernetes/cmd/controller/options. investigate consequences
	crds := []struct {
		name         string
		plural       string
		objType      runtime.Object
		resyncPeriod time.Duration
	}{
		{
			name:         "component-build",
			plural:       crv1.ComponentBuildResourcePlural,
			objType:      &crv1.ComponentBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "config",
			plural:       crv1.ConfigResourcePlural,
			objType:      &crv1.Config{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "service-build",
			plural:       crv1.ServiceBuildResourcePlural,
			objType:      &crv1.ServiceBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "system-build",
			plural:       crv1.SystemBuildResourcePlural,
			objType:      &crv1.SystemBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
	}

	informersMap := map[string]cache.SharedInformer{}
	for _, crd := range crds {
		listerWatcher := cache.NewListWatchFromClient(
			latticeResourceClient,
			crd.plural,
			apiv1.NamespaceAll,
			fields.Everything(),
		)
		informer := cache.NewSharedInformer(
			listerWatcher,
			crd.objType,
			crd.resyncPeriod,
		)
		informersMap[crd.name] = informer
	}

	return informersMap
}

type controllerInitializer func(context ControllerContext)

func CreateControllerInitializers() map[string]controllerInitializer {
	return map[string]controllerInitializer{
		"component-build": initializeComponentBuildController,
		"service-build":   initializeServiceBuildController,
		"system-build":    initializeSystemBuildController,
	}
}

func StartControllers(ctx ControllerContext, initializers map[string]controllerInitializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
