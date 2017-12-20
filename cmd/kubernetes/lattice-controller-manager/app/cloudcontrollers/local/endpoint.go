package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/cloud/local/endpoint"
)

func initializeEndpointController(ctx controller.Context) {
	go endpoint.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-local-endpoints"),
		ctx.LatticeInformerFactory.Lattice().V1().Endpoints(),
	).Run(4, ctx.Stop)
}
