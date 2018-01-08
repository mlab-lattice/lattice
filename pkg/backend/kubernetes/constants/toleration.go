package constants

import (
	corev1 "k8s.io/api/core/v1"
)

var TolerateAllTaints corev1.Toleration = corev1.Toleration{
	Operator: corev1.TolerationOpExists,
}

var TolerationMasterNode corev1.Toleration = corev1.Toleration{
	Key:      LabelKeyMasterNode,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationBuildNode corev1.Toleration = corev1.Toleration{
	Key:      LabelKeyBuildNode,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationNodePool corev1.Toleration = corev1.Toleration{
	Key:      LabelKeyNodeRoleNodePool,
	Operator: corev1.TolerationOpExists,
	Effect:   corev1.TaintEffectNoSchedule,
}
