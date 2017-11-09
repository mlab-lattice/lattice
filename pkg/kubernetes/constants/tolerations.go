package constants

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	TaintMasterNode       = "node-role.kubernetes.io/master"
	TaintBuildNode        = "node-role.lattice.mlab.com/build"
	TaintLatticeNamespace = "namespace.lattice.mlab.com"
)

var TolerateAllTaints corev1.Toleration = corev1.Toleration{
	Operator: corev1.TolerationOpExists,
}

var TolerationMasterNode corev1.Toleration = corev1.Toleration{
	Key:      TaintMasterNode,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}

var TolerationBuildNode corev1.Toleration = corev1.Toleration{
	Key:      TaintBuildNode,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}
