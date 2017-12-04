package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/kubernetes/componentbuild"
)

func initializeComponentBuildController(ctx controller.Context) {
	go componentbuild.NewController(
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.CRInformers.Config,
		ctx.CRInformers.ComponentBuild,
		ctx.InformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
