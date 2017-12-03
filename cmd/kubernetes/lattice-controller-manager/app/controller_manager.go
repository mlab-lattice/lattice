package app

import (
	"time"

	awscontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/aws"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/kubernetescontrollers"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/latticecontrollers"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

func Run(kubeconfig, provider, terraformModulePath string) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	latticeResourceClient, _, err := customresource.NewClient(config)
	if err != nil {
		panic(err)
	}

	// TODO: setting stop as nil for now, won't actually need it until leader-election is used
	ctx := CreateControllerContext(rest.Interface(latticeResourceClient), config, nil, terraformModulePath)
	glog.V(1).Info("Starting controllers")
	StartControllers(ctx, GetControllerInitializers(provider))

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
	stop <-chan struct{},
	terraformModulePath string,
) controller.Context {
	cb := controller.ClientBuilder{
		Kubeconfig: kubeconfig,
	}

	versionedClient := cb.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactory(versionedClient, time.Duration(12*time.Hour))

	return controller.Context{
		InformerFactory:           sharedInformers,
		CRDInformers:              getCRDInformers(latticeResourceClient),
		LatticeResourceRestClient: latticeResourceClient,
		ClientBuilder:             cb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
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
			plural:       crv1.ResourcePluralComponentBuild,
			objType:      &crv1.ComponentBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "config",
			plural:       crv1.ResourcePluralConfig,
			objType:      &crv1.Config{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "service",
			plural:       crv1.ResourcePluralService,
			objType:      &crv1.Service{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "service-build",
			plural:       crv1.ResourcePluralServiceBuild,
			objType:      &crv1.ServiceBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "system",
			plural:       crv1.ResourcePluralSystem,
			objType:      &crv1.System{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "system-build",
			plural:       crv1.ResourcePluralSystemBuild,
			objType:      &crv1.SystemBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "system-rollout",
			plural:       crv1.ResourcePluralSystemRollout,
			objType:      &crv1.SystemRollout{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			name:         "system-teardown",
			plural:       crv1.ResourcePluralSystemTeardown,
			objType:      &crv1.SystemTeardown{},
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

func GetControllerInitializers(provider string) map[string]controller.Initializer {
	initializers := map[string]controller.Initializer{}

	for name, initializer := range kubernetescontrollers.GetControllerInitializers() {
		initializers["kubernetes-"+name] = initializer
	}

	for name, initializer := range latticecontrollers.GetControllerInitializers() {
		initializers["lattice-"+name] = initializer
	}

	switch provider {
	case constants.ProviderAWS:
		for name, initializer := range awscontrollers.GetControllerInitializers() {
			initializers["cloud-aws-"+name] = initializer
		}
	case constants.ProviderLocal:
		// Local case doesn't need any extra controllers
	default:
		panic("unsupported provider " + provider)
	}

	return initializers
}

func StartControllers(ctx controller.Context, initializers map[string]controller.Initializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
