package kubernetes

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"
)

func NodePoolToleration(nodePool *crv1.NodePool) corev1.Toleration {
	return corev1.Toleration{
		Key:      constants.LabelKeyNodeRoleNodePool,
		Operator: corev1.TolerationOpEqual,
		Value:    NodePoolIDLabelValue(nodePool),
		Effect:   corev1.TaintEffectNoSchedule,
	}
}
