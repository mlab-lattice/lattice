package kubernetes

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	corev1 "k8s.io/api/core/v1"
)

func GetServiceTaintToleration(svcName string) corev1.Toleration {
	return corev1.Toleration{
		Key:      constants.TaintServiceNode,
		Operator: corev1.TolerationOpEqual,
		Value:    svcName,
		Effect:   corev1.TaintEffectNoSchedule,
	}
}
