package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/cloud/aws/loadbalancer"
)

func initializeLoadBalancerController(ctx controller.Context) {
	awsCloudProvider := ctx.CloudProvider.(*aws.DefaultAWSCloudProvider)

	go loadbalancer.NewController(
		ctx.ClusterID,
		aws.CloudProvider(awsCloudProvider),
		ctx.TerraformModulePath,
		ctx.TerraformBackendOptions,
		ctx.KubeClientBuilder.ClientOrDie("lattice-controller-aws-load-balancer"),
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-aws-load-balancer"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().LoadBalancers(),
		ctx.LatticeInformerFactory.Lattice().V1().NodePools(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.KubeInformerFactory.Core().V1().Services(),
	).Run(4, ctx.Stop)
}
