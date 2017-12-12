package app

import (
	"time"

	awscontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/aws"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/kubernetescontrollers"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/latticecontrollers"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/constants"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
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

	glog.V(4).Info("Starting informer factory kubeinformers")
	ctx.KubeInformerFactory.Start(ctx.Stop)
	ctx.LatticeInformerFactory.Start(ctx.Stop)

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

	versionedKubeClient := kcb.ClientOrDie("shared-kubeinformers")
	kubeInformers := kubeinformers.NewSharedInformerFactory(versionedKubeClient, time.Duration(12*time.Hour))

	versionedLatticeClient := lcb.ClientOrDie("shared-latticeinformers")
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	return controller.Context{
		KubeInformerFactory:    kubeInformers,
		LatticeInformerFactory: latticeInformers,
		KubeClientBuilder:      kcb,
		LatticeClientBuilder:   lcb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
	}
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
