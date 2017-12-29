package constants

import (
	corev1 "k8s.io/api/core/v1"
)

var NodeAffinityMasterNode = corev1.NodeAffinity{
	RequiredDuringSchedulingIgnoredDuringExecution: &NodeSelectorMasterNode,
}
