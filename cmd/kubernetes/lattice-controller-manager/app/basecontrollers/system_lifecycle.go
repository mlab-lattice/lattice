package basecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/systemlifecycle"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemlifecycle.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-lifecycle"),
		ctx.LatticeInformerFactory.Lattice().V1().Deploies(),
		ctx.LatticeInformerFactory.Lattice().V1().Teardowns(),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Builds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
	).Run(4, ctx.Stop)
}
