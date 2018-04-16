package constants

import (
	corev1 "k8s.io/api/core/v1"
)

var NodeSelectorMasterNode = corev1.NodeSelector{
	NodeSelectorTerms: []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      LabelKeyNodeRoleLatticeMaster,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	},
}

var NodeSelectorBuildNode = corev1.NodeSelector{
	NodeSelectorTerms: []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      LabelKeyNodeRoleLatticeBuild,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	},
}

var NodeSelectorNodePool = corev1.NodeSelector{
	NodeSelectorTerms: []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      LabelKeyNodeRoleLatticeNodePool,
					Operator: corev1.NodeSelectorOpExists,
				},
			},
		},
	},
}
