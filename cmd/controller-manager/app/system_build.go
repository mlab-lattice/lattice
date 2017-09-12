package app

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/systembuild"
)

func initializeSystemBuildController(ctx ControllerContext) {
	go systembuild.NewSystemBuildController(
		ctx.Provider,
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-build"],
		ctx.CRDInformers["service-build"],
	).Run(4, ctx.Stop)
}
