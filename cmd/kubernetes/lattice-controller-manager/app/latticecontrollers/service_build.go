package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/servicebuild"
)

func initializeServiceBuildController(ctx controller.Context) {
	go servicebuild.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-service"),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
	).Run(4, ctx.Stop)
}
