package app

import (
	"time"

	awscontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/aws"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/kubernetescontrollers"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/latticecontrollers"
	"github.com/mlab-lattice/system/pkg/constants"
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

	// TODO: setting stop as nil for now, won't actually need it until leader-election is used
	ctx := CreateControllerContext(config, nil, terraformModulePath)
	glog.V(1).Info("Starting controllers")
	StartControllers(ctx, GetControllerInitializers(provider))

	glog.V(4).Info("Starting informer factory informers")
	ctx.InformerFactory.Start(ctx.Stop)
	ctx.CRInformers.Start(ctx.Stop)

	select {}
}

func CreateControllerContext(
	kubeconfig *rest.Config,
	stop <-chan struct{},
	terraformModulePath string,
) controller.Context {
	kcb := controller.KubeClientBuilder{
		Kubeconfig: kubeconfig,
	}
	lcb := controller.LatticeClientBuilder{
		Kubeconfig: kubeconfig,
	}

	versionedKubeClient := kcb.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactory(versionedKubeClient, time.Duration(12*time.Hour))

	latticeClient := lcb.ClientOrDie("shared-informers")

	return controller.Context{
		InformerFactory:      sharedInformers,
		CRInformers:          getCRDInformers(latticeClient.V1().RESTClient()),
		KubeClientBuilder:    kcb,
		LatticeClientBuilder: lcb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
	}
}

func getCRDInformers(latticeResourceClient rest.Interface) *controller.CRInformers {
	crdInformers := &controller.CRInformers{}
	// FIXME: defaultResync blindly taken from k8s.io/kubernetes/cmd/controller/options. investigate consequences
	crds := []struct {
		dest         *cache.SharedInformer
		plural       string
		objType      runtime.Object
		resyncPeriod time.Duration
	}{
		{
			dest:         &crdInformers.ComponentBuild,
			plural:       crv1.ResourcePluralComponentBuild,
			objType:      &crv1.ComponentBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.Config,
			plural:       crv1.ResourcePluralConfig,
			objType:      &crv1.Config{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.Service,
			plural:       crv1.ResourcePluralService,
			objType:      &crv1.Service{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.ServiceBuild,
			plural:       crv1.ResourcePluralServiceBuild,
			objType:      &crv1.ServiceBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.System,
			plural:       crv1.ResourcePluralSystem,
			objType:      &crv1.System{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.SystemBuild,
			plural:       crv1.ResourcePluralSystemBuild,
			objType:      &crv1.SystemBuild{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.SystemRollout,
			plural:       crv1.ResourcePluralSystemRollout,
			objType:      &crv1.SystemRollout{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
		{
			dest:         &crdInformers.SystemTeardown,
			plural:       crv1.ResourcePluralSystemTeardown,
			objType:      &crv1.SystemTeardown{},
			resyncPeriod: time.Duration(12 * time.Hour),
		},
	}

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
		*crd.dest = informer
	}

	return crdInformers
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
