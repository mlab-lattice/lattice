package aws

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/cloud/aws/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewController(
		ctx.KubeClientBuilder.ClientOrDie("lattice-controller-cloud-aws-service"),
		ctx.LatticeClientBuilder.ClientOrDie("lattice-controller-cloud-aws-service"),
		ctx.CRInformers.Config,
		ctx.CRInformers.Service,
		ctx.InformerFactory.Core().V1().Services(),
		ctx.TerraformModulePath,
	).Run(128, ctx.Stop)
}
