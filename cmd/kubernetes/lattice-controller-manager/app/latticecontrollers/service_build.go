package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/servicebuild"
)

func initializeServiceBuildController(ctx controller.Context) {
	go servicebuild.NewServiceBuildController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
	).Run(4, ctx.Stop)
}
