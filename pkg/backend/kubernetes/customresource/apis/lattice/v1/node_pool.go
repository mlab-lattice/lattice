package v1

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
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

func (np *NodePool) ID(epoch int64) string {
	return fmt.Sprintf("%v.%v.%v", np.Namespace, np.Name, epoch)
}

func (np *NodePool) Description() string {
	// TODO: when adding lattice node pools may have to adjust his
	systemID, err := kubeutil.SystemID(np.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", np.Namespace))
	}

	typeDescription := "UNKNOWN"
	if np.Labels != nil {
		if serviceID, ok := np.Labels[constants.LabelKeyServiceID]; ok {
			typeDescription = fmt.Sprintf("dedicated for service %v", serviceID)
		}

		// FIXME: add path type for system node pools
	}

	return fmt.Sprintf("node pool %v (%v in system %v)", np.Name, typeDescription, systemID)
}

func (np *NodePool) CurrentEpoch() (int64, bool) {
	if len(np.Status.Epochs) == 0 {
		return 0, false
	}

	var epochs []int64
	for epoch := range np.Status.Epochs {
		epochs = append(epochs, epoch)
	}

	sort.Slice(epochs, func(i, j int) bool {
		return epochs[i] < epochs[j]
	})

	return epochs[len(epochs)-1], true
}

func (np *NodePool) Affinity(epoch int64) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      constants.LabelKeyNodeRoleLatticeNodePool,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{np.ID(epoch)},
						},
					},
				},
			},
		},
	}
}

func (np *NodePool) Toleration(epoch int64) corev1.Toleration {
	return corev1.Toleration{
		Key:      constants.LabelKeyNodeRoleLatticeNodePool,
		Operator: corev1.TolerationOpEqual,
		Value:    np.ID(epoch),
		Effect:   corev1.TaintEffectNoSchedule,
	}
}

type NodePoolSpec struct {
	NumInstances int32  `json:"numInstances"`
	InstanceType string `json:"instanceType"`
}

type NodePoolStatus struct {
	ObservedGeneration int64                         `json:"observedGeneration"`
	State              NodePoolState                 `json:"state"`
	Epochs             map[int64]NodePoolStatusEpoch `json:"epochs"`
}

type NodePoolState string

const (
	NodePoolStatePending  NodePoolState = "pending"
	NodePoolStateScaling  NodePoolState = "scaling"
	NodePoolStateUpdating NodePoolState = "updating"
	NodePoolStateStable   NodePoolState = "stable"
	NodePoolStateFailed   NodePoolState = "failed"
)

type NodePoolStatusEpoch struct {
	Spec  NodePoolSpec  `json:"spec"`
	State NodePoolState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NodePool `json:"items"`
}
