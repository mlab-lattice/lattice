package app

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/componentbuild"
)

func initializeBuildController(ctx ControllerContext) {
	go componentbuild.NewComponentBuildController(
		ctx.Provider,
		ctx.ClientBuilder.ClientOrDie("build-controller"),
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["config"],
		ctx.CRDInformers["build"],
		ctx.InformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
