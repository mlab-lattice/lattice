package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/system"
)

func initializeSystemController(ctx controller.Context) {
	go system.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system"),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
	).Run(4, ctx.Stop)
}
