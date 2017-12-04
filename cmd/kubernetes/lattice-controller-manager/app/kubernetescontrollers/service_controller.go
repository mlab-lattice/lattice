package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/kubernetes/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewController(
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.CRInformers.Config,
		ctx.CRInformers.System,
		ctx.CRInformers.Service,
		ctx.InformerFactory.Apps().V1beta2().Deployments(),
		ctx.InformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
