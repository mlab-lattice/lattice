package kubernetes

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"
)

func NodePoolNodeAffinity(nodePool *crv1.NodePool) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      constants.LabelKeyNodeRoleNodePool,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{NodePoolIDLabelValue(nodePool)},
						},
					},
				},
			},
		},
	}
}
