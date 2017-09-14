package latticecontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/lattice/systemrollout"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemrollout.NewSystemRolloutController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-rollout"],
		ctx.CRDInformers["system"],
		ctx.CRDInformers["system-build"],
	).Run(4, ctx.Stop)
}
