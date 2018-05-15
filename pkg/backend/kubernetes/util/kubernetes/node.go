package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

func NumReadyNodes(nodes []corev1.Node) int32 {
	var ready int32
	for _, node := range nodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					ready++
				}

				continue
			}
		}
	}

	return ready
}
