package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/cloud/aws/endpoint"
)

func initializeEndpointController(ctx controller.Context) {
	awsCloudProvider := ctx.CloudProvider.(*aws.DefaultAWSCloudProvider)

	go endpoint.NewController(
		ctx.ClusterID,
		aws.CloudProvider(awsCloudProvider),
		ctx.TerraformModulePath,
		ctx.TerraformBackendOptions,
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-aws-endpoints"),
		ctx.LatticeInformerFactory.Lattice().V1().Endpoints(),
	).Run(4, ctx.Stop)
}
