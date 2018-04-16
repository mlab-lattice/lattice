package constants

import (
	corev1 "k8s.io/api/core/v1"
)

var TolerationKubernetesMasterNode = corev1.Toleration{
	Key:      LabelKeyNodeRollKubernetesMaster,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationLatticeMasterNode = corev1.Toleration{
	Key:      LabelKeyNodeRoleLatticeMaster,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationBuildNode = corev1.Toleration{
	Key:      LabelKeyNodeRoleLatticeBuild,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationNodePool = corev1.Toleration{
	Key:      LabelKeyNodeRoleLatticeNodePool,
	Operator: corev1.TolerationOpExists,
	Effect:   corev1.TaintEffectNoSchedule,
}
