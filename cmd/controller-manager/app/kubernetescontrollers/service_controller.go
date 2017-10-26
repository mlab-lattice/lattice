package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/kubernetes/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewServiceController(
		ctx.Provider,
		ctx.ClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["config"],
		ctx.CRDInformers["system"],
		ctx.CRDInformers["service"],
		ctx.InformerFactory.Apps().V1beta2().Deployments(),
		ctx.InformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
