package main

import (
	"time"

	localcontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-dns-local/localcontrollers"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/kubernetescontrollers"
	controllermanager "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app"
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

	ctx := controllermanager.CreateControllerContext(config, nil, terraformModulePath)

	initializers := map[string]controller.Initializer{}

	switch provider {
	case constants.ProviderLocal:
		for name, initializer := range localcontrollers.GetControllerInitializers() {
			initializers["local-"+name] = initializer
		}
	default:
		panic("lattice-local-dns is only supported for local provider.")
	}

	glog.V(1).Info("Starting controllers")
	StartControllers(ctx, initializers)

	glog.V(4).Info("Starting informer factory")
	latticeinformers.Start(ctx.Stop)

	select {}
}

func StartControllers(ctx controller.Context, initializers map[string]controller.Initializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
