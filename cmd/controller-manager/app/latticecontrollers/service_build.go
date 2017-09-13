package latticecontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/lattice/servicebuild"
)

func initializeServiceBuildController(ctx controller.Context) {
	go servicebuild.NewServiceBuildController(
		ctx.Provider,
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
	).Run(4, ctx.Stop)
}
