package aws

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewAWSCloudProvider() *DefaultLocalCloudProvider {
	return &DefaultLocalCloudProvider{}
}

type DefaultLocalCloudProvider struct {
}

func (cp *DefaultLocalCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	// nothing to do
}

func (cp *DefaultLocalCloudProvider) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	// nothing to do
}

func (cp *DefaultLocalCloudProvider) TransformPodTemplateSpec(template *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	// nothing to do
	return template
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	// nothing to do
	return spec
}

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	// nothing to do
	return spec
}

func (cp *DefaultLocalCloudProvider) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// nothing to do
	return true, "", current
}
