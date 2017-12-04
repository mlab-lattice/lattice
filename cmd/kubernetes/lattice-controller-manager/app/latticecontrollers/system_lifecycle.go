package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/systemlifecycle"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemlifecycle.NewController(
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-lattice-system-lifecycle"),
		ctx.CRInformers.SystemRollout,
		ctx.CRInformers.SystemTeardown,
		ctx.CRInformers.System,
		ctx.CRInformers.SystemBuild,
		ctx.CRInformers.ServiceBuild,
		ctx.CRInformers.ComponentBuild,
	).Run(4, ctx.Stop)
}
