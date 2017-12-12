package localcontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/kubernetes/local"
)

func initialiseAddressController(ctx controller.Context) {
	go local.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-build"),
		// To become an Address informer?
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
	).Run(4, ctx.Stop)
}
