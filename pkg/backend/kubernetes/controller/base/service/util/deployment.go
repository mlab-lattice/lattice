package util

import (
	appsv1 "k8s.io/api/apps/v1"

	k8sappsv1 "k8s.io/kubernetes/pkg/apis/apps/v1"
)

func SetDeploymentSpecDefaults(spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	// Copy so the shared cache isn't mutated
	spec = spec.DeepCopy()

	deployment := &appsv1.Deployment{
		Spec: *spec,
	}
	k8sappsv1.SetObjectDefaults_Deployment(deployment)

	return &deployment.Spec
}
