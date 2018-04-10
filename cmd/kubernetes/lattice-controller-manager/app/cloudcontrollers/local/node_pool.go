package aws

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/cloud/local/nodepool"
)

func initializeNodePoolController(ctx controller.Context) {
	go nodepool.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-local-node-pool"),
		ctx.LatticeInformerFactory.Lattice().V1().NodePools(),
	).Run(4, ctx.Stop)
}
