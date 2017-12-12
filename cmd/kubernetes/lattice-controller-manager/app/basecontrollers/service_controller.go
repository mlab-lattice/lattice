package basecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewController(
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.KubeInformerFactory.Apps().V1beta2().Deployments(),
		ctx.KubeInformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
