package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/serviceaddress"
)

func initializeServiceAddressController(ctx controller.Context) {
	go serviceaddress.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-service-address"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceAddresses(),
		ctx.LatticeInformerFactory.Lattice().V1().Endpoints(),
	).Run(4, ctx.Stop)
}
