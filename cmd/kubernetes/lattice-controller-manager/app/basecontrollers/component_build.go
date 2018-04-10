package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/componentbuild"
)

func initializeComponentBuildController(ctx controller.Context) {
	go componentbuild.NewController(
		ctx.LatticeID,
		ctx.CloudProvider,
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-build-controller"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
		ctx.KubeInformerFactory.Batch().V1().Jobs(),
	).Run(4, ctx.Stop)
}
