package app

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/componentbuild"
)

func initializeBuildController(ctx ControllerContext) {
	go componentbuild.NewComponentBuildController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["build"],
		ctx.CRDInformers["config"],
		ctx.InformerFactory.Batch().V1().Jobs(),
		ctx.ClientBuilder.ClientOrDie("build-controller"),
		ctx.Provider,
	).Run(4, ctx.Stop)
}
