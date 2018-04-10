package aws

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/cloud/local/loadbalancer"
)

func initializeLoadBalancerController(ctx controller.Context) {
	localCloudProvider := ctx.CloudProvider.(*local.DefaultLocalCloudProvider)

	go loadbalancer.NewController(
		ctx.KubeClientBuilder.ClientOrDie("load-balancer"),
		ctx.LatticeClientBuilder.ClientOrDie("load-balancer"),
		local.CloudProvider(localCloudProvider),
		ctx.LatticeInformerFactory.Lattice().V1().LoadBalancers(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.KubeInformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
