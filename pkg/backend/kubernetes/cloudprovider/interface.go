package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
    "github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	// TransformComponentBuildJobSpec takes in the JobSpec generated for a ComponentBuild, and applies any cloud provider
	// related transforms necessary to a copy of the JobSpec, and returns it.
	TransformComponentBuildJobSpec(*batchv1.JobSpec) *batchv1.JobSpec

	ComponentBuildWorkDirectoryVolumeSource(jobName string) corev1.VolumeSource

	// TransformServiceDeploymentSpec takes in the DeploymentSpec generated for a Service, and applies any cloud provider
	// related transforms necessary to a copy of the DeploymentSpec, and returns it.
	TransformServiceDeploymentSpec(*crv1.Service, *appsv1.DeploymentSpec) *appsv1.DeploymentSpec

	// IsDeploymentSpecCurrent checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformServiceDeploymentSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformServiceDeploymentSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformServiceDeploymentSpec.
	IsDeploymentSpecUpdated(service *crv1.Service, current, desired, untransformed *appsv1.DeploymentSpec) (bool, string, *appsv1.DeploymentSpec)
}

type Options struct {
	AWS   *aws.Options
	Local *local.Options
}

func NewCloudProvider(clusterID types.ClusterID, options *Options) (Interface, error) {
	if options.AWS != nil {
		return aws.NewCloudProvider(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewCloudProvider(clusterID, options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}
