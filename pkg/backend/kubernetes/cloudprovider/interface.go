package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/constants"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

type Interface interface {
	TransformComponentBuildJobSpec(*batchv1.JobSpec) *batchv1.JobSpec
	TransformServiceDeploymentSpec(*appsv1.DeploymentSpec) *appsv1.DeploymentSpec
}

func NewCloudProvider(providerName string) (Interface, error) {
	switch providerName {
	case constants.ProviderLocal:
		return local.NewLocalCloudProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %v", providerName)
	}
}
