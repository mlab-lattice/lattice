package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/systemlifecycle"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemlifecycle.NewController(
		ctx.LatticeID,
		ctx.KubeClientBuilder.ClientOrDie("lattice-controller-lattice-system-lifecycle"),
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-lifecycle"),
		ctx.LatticeInformerFactory.Lattice().V1().Deploies(),
		ctx.LatticeInformerFactory.Lattice().V1().Teardowns(),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Builds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
	).Run(4, ctx.Stop)
}
