package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/equality"
)

func ContainersSemanticallyEqual(container1, container2 *corev1.Container) bool {
	c1Copy := container1.DeepCopy()
	c2Copy := container2.DeepCopy()

	return equality.Semantic.DeepEqual(c1Copy, c2Copy)
}

func PodTemplateSpecsSemanticallyEqual(template1, template2 *corev1.PodTemplateSpec) bool {
	// Below gleamed from: https://github.com/kubernetes/kubernetes/blob/v1.9.0/pkg/controller/deployment/util/deployment_util.go#L634-L655
	t1Copy := template1.DeepCopy()
	t2Copy := template2.DeepCopy()

	// First, compare template.Labels (ignoring hash)
	labels1, labels2 := t1Copy.Labels, t2Copy.Labels
	if len(labels1) > len(labels2) {
		labels1, labels2 = labels2, labels1
	}
	// We make sure len(labels2) >= len(labels1)
	for k, v := range labels2 {
		if labels1[k] != v && k != extensionsv1beta1.DefaultDeploymentUniqueLabelKey {
			return false
		}
	}
	// Then, compare the templates without comparing their labels
	t1Copy.Labels, t2Copy.Labels = nil, nil
	return equality.Semantic.DeepEqual(t1Copy, t2Copy)
}

func VolumesSemanticallyEqual(volume1, volume2 *corev1.Volume) bool {
	v1Copy := volume1.DeepCopy()
	v2Copy := volume2.DeepCopy()

	return equality.Semantic.DeepEqual(v1Copy, v2Copy)
}
