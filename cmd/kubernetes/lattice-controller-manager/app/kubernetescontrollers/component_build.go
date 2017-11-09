package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/kubernetes/componentbuild"
)

func initializeComponentBuildController(ctx controller.Context) {
	go componentbuild.NewComponentBuildController(
		ctx.ClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["config"],
		ctx.CRDInformers["component-build"],
		ctx.InformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
