package app

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/servicebuild"
)

func initializeServiceBuildController(ctx ControllerContext) {
	go servicebuild.NewServiceBuildController(
		ctx.Provider,
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
	).Run(4, ctx.Stop)
}
