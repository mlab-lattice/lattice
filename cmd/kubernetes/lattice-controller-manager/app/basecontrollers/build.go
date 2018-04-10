package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/build"
)

func initializeSystemBuildController(ctx controller.Context) {
	go build.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-build"),
		ctx.LatticeInformerFactory.Lattice().V1().Builds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
	).Run(4, ctx.Stop)
}
