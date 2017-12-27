package main

import (
	"time"

	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	dnscontroller "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local/controller"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

func Run(clusterIDString, kubeconfig, provider, terraformModulePath string,
	serverConfigPath string, hostConfigPath string) {

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

	clusterID := types.ClusterID(clusterIDString)

	ctx, err := CreateControllerContext(clusterID, config, nil, terraformModulePath)

	if err != nil {
		panic(err)
	}

	if provider != constants.ProviderLocal {
		panic("lattice-local-dns is only supported for local provider.")
	}

	glog.V(1).Info("Starting dns controller")

	go dnscontroller.NewController(
		serverConfigPath,
		hostConfigPath,
		ctx.LatticeClientBuilder.ClientOrDie("local-dns-lattice-address"),
		ctx.LatticeInformerFactory.Lattice().V1().Endpoints(),
	).Run(4, ctx.Stop)

	glog.V(1).Info("Starting informer factory")
	ctx.LatticeInformerFactory.Start(ctx.Stop)

	select {}
}

func CreateControllerContext(
	clusterID types.ClusterID,
	kubeconfig *rest.Config,
	stop <-chan struct{},
	terraformModulePath string,
) (controller.Context, error) {

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

	ctx := controller.Context{
		ClusterID: clusterID,

		KubeInformerFactory:    kubeInformers,
		LatticeInformerFactory: latticeInformers,
		KubeClientBuilder:      kcb,
		LatticeClientBuilder:   lcb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
	}
	return ctx, nil
}
