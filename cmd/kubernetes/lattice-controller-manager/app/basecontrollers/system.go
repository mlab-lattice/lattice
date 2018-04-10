package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/base/system"
)

func initializeSystemController(ctx controller.Context) {
	go system.NewController(
		ctx.LatticeID,
		ctx.SystemBootstrappers,
		ctx.KubeClientBuilder.ClientOrDie("lattice-controller-lattice-system"),
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.KubeInformerFactory.Core().V1().Namespaces(),
	).Run(4, ctx.Stop)
}
