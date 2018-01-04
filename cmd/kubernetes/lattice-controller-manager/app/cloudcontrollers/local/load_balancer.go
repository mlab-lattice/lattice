package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/cloud/local/loadbalancer"
)

func initializeLoadBalancerController(ctx controller.Context) {
	go loadbalancer.NewController(
		ctx.KubeClientBuilder.ClientOrDie("load-balancer"),
		ctx.LatticeClientBuilder.ClientOrDie("load-balancer"),
		ctx.LatticeInformerFactory.Lattice().V1().LoadBalancers(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.KubeInformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
