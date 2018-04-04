package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularNodePool  = "nodepool"
	ResourcePluralNodePool    = "nodepools"
	ResourceShortNameNodePool = "lnp"
	ResourceScopeNodePool     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NodePoolSpec   `json:"spec"`
	Status            NodePoolStatus `json:"status,omitempty"`
}

func (np *NodePool) IDLabelValue() string {
	return fmt.Sprintf("%v.%v", np.Namespace, np.Name)
}

func (np *NodePool) Affinity() *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      constants.LabelKeyNodeRoleNodePool,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{np.IDLabelValue()},
						},
					},
				},
			},
		},
	}
}

func (np *NodePool) Toleration() corev1.Toleration {
	return corev1.Toleration{
		Key:      constants.LabelKeyNodeRoleNodePool,
		Operator: corev1.TolerationOpEqual,
		Value:    np.IDLabelValue(),
		Effect:   corev1.TaintEffectNoSchedule,
	}
}

type NodePoolSpec struct {
	NumInstances int32  `json:"numInstances"`
	InstanceType string `json:"instanceType"`
}

type NodePoolStatus struct {
	State NodePoolState `json:"state"`
}

type NodePoolState string

const (
	NodePoolStatePending     NodePoolState = "pending"
	NodePoolStateScalingDown NodePoolState = "scaling down"
	NodePoolStateScalingUp   NodePoolState = "scaling up"
	NodePoolStateStable      NodePoolState = "stable"
	NodePoolStateFailed      NodePoolState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NodePool `json:"items"`
}
