package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/systembuild"
)

func initializeSystemBuildController(ctx controller.Context) {
	go systembuild.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-build"),
		ctx.LatticeInformerFactory.Lattice().V1().SystemBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
	).Run(4, ctx.Stop)
}
