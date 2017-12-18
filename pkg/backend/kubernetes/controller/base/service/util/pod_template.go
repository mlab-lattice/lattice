package util

import (
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/equality"
)

// Inspired by https://github.com/kubernetes/kubernetes/blob/fc8bfe2d8929e11a898c4557f9323c482b5e8842/pkg/controller/deployment/util/deployment_util.go#L634-L655
func PodTemplateSpecsSemanticallyEqual(template1, template2 corev1.PodTemplateSpec) bool {
	t1Copy := template1.DeepCopy()
	t2Copy := template2.DeepCopy()

	// First, compare template.Labels (ignoring hash)
	labels1, labels2 := t1Copy.Labels, t2Copy.Labels
	if len(labels1) > len(labels2) {
		labels1, labels2 = labels2, labels1
	}

	// Make sure len(labels2) >= len(labels1)
	for k, v := range labels2 {
		if labels1[k] != v && k != extensionsv1beta1.DefaultDeploymentUniqueLabelKey {
			return false
		}
	}

	// Compare the templates without comparing their labels
	t1Copy.Labels, t2Copy.Labels = nil, nil
	return equality.Semantic.DeepEqual(t1Copy, t2Copy)
}
