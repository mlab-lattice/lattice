package local

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

func NewLocalCloudProvider() *DefaultLocalCloudProvider {
	return &DefaultLocalCloudProvider{}
}

type DefaultLocalCloudProvider struct {
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	spec.Template.Spec.Affinity = nil
	return spec
}

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	spec.Template.Spec.Affinity = nil
	return spec
}
