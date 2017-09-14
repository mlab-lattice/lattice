package latticecontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/lattice/system"
)

func initializeSystemController(ctx controller.Context) {
	go system.NewSystemController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system"],
		ctx.CRDInformers["service"],
	).Run(4, ctx.Stop)
}
