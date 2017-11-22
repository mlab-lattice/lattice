package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/systemlifecycle"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemlifecycle.NewSystemLifecycleController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-rollout"],
		ctx.CRDInformers["system-teardown"],
		ctx.CRDInformers["system"],
		ctx.CRDInformers["system-build"],
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
	).Run(4, ctx.Stop)
}
