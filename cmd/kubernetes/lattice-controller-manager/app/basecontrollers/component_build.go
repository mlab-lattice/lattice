package basecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/componentbuild"
)

func initializeComponentBuildController(ctx controller.Context) {
	go componentbuild.NewController(
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
		ctx.KubeInformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
