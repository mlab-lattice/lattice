package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/servicebuild"
)

func initializeServiceBuildController(ctx controller.Context) {
	go servicebuild.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-service"),
		ctx.CRInformers.ServiceBuild,
		ctx.CRInformers.ComponentBuild,
	).Run(4, ctx.Stop)
}
