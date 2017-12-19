package kubernetes

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"

	k8scorev1 "k8s.io/kubernetes/pkg/apis/core/v1"
)

func PodTemplatesEqual(a, b *corev1.PodTemplate) bool {
	k8scorev1.SetObjectDefaults_PodTemplate(a)
	k8scorev1.SetObjectDefaults_PodTemplate(b)

	return reflect.DeepEqual(a, b)
}
