package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/kubernetes/componentbuild"
)

func initializeComponentBuildController(ctx controller.Context) {
	go componentbuild.NewComponentBuildController(
		ctx.Provider,
		ctx.ClientBuilder.ClientOrDie("build-controller"),
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["config"],
		ctx.CRDInformers["component-build"],
		ctx.InformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
