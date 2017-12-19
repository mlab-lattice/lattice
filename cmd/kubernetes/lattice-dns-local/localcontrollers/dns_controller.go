package localcontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/kubernetes/local"
)

func initialiseAddressController(ctx controller.Context) {
	go local.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("local-dns-lattice-address"),
		// To become an Address informer, now a System informer for debugging.
		ctx.LatticeInformerFactory.Lattice().V1().SystemBuilds(),
	).Run(4, ctx.Stop)
}
