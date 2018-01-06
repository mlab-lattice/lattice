package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/cloud/aws/nodepool"
)

func initializeNodePoolController(ctx controller.Context) {
	awsCloudProvider := ctx.CloudProvider.(*aws.DefaultAWSCloudProvider)

	go nodepool.NewController(
		ctx.ClusterID,
		aws.CloudProvider(awsCloudProvider),
		ctx.TerraformModulePath,
		ctx.TerraformBackendOptions,
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-aws-endpoints"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().NodePools(),
	).Run(4, ctx.Stop)
}
