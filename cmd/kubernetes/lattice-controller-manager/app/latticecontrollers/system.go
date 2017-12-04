package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/system"
)

func initializeSystemController(ctx controller.Context) {
	go system.NewSystemController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system"),
		ctx.CRInformers.System,
		ctx.CRInformers.Service,
	).Run(4, ctx.Stop)
}
