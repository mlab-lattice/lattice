package constants

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	TaintMasterNode = "node-role.kubernetes.io/master"
)

var TolerationMasterNode corev1.Toleration = corev1.Toleration{
	Key:      TaintMasterNode,
	Operator: corev1.TolerationOpEqual,
	Value:    "true",
	Effect:   corev1.TaintEffectNoSchedule,
}
