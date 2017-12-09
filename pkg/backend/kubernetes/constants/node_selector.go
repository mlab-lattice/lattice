package constants

import (
	corev1 "k8s.io/api/core/v1"
)

var NodeSelectorMasterNode = corev1.NodeSelector{
	NodeSelectorTerms: []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      LabelKeyMasterNode,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	},
}
