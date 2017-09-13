package app

import (
	"time"

	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/kubernetescontrollers"
	"github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/latticecontrollers"
	latticeresource "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/informers"
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
	StartControllers(ctx, GetControllerInitializers())

	glog.V(4).Info("Starting informer factory informers")
	ctx.InformerFactory.Start(ctx.Stop)

	for resource, informer := range ctx.CRDInformers {
		glog.V(4).Infof("Starting %v informer", resource)
		go informer.Run(ctx.Stop)

	}

	select {}
}

func CreateControllerContext(
	latticeResourceClient rest.Interface,
	kubeconfig *rest.Config,
	provider string,
	stop <-chan struct{},
) controller.Context {
	cb := controller.ClientBuilder{
		Kubeconfig: kubeconfig,
	}

	versionedClient := cb.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactory(versionedClient, time.Duration(12*time.Hour))

	return controller.Context{
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

func GetControllerInitializers() map[string]controller.Initializer {
	initializers := map[string]controller.Initializer{}

	for name, initializer := range kubernetescontrollers.GetControllerInitializers() {
		initializers["kubernetes-"+name] = initializer
	}

	for name, initializer := range latticecontrollers.GetControllerInitializers() {
		initializers["lattice-"+name] = initializer
	}

	return initializers
}

func StartControllers(ctx controller.Context, initializers map[string]controller.Initializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
