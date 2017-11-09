package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/systembuild"
)

func initializeSystemBuildController(ctx controller.Context) {
	go systembuild.NewSystemBuildController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-build"],
		ctx.CRDInformers["service-build"],
	).Run(4, ctx.Stop)
}
