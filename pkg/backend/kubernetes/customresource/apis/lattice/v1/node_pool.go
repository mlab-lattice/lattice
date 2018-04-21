package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

const (
	ResourceSingularNodePool = "nodepool"
	ResourcePluralNodePool   = "nodepools"
	ResourceScopeNodePool    = apiextensionsv1beta1.NamespaceScoped
)

// NodePoolServiceDedicatedID is the key for a label indicating that the node pool is dedicated for a service.
// The label's value should be the ID of the service.
var NodePoolServiceDedicatedIDLabelKey = fmt.Sprintf("service.dedicated.node-pool.%v/id", SchemeGroupVersion.String())

// NodePoolServiceDedicatedID is the key for a label indicating that the node pool is shared for a system.
// The label's value should be the node pool's path in the system definition.
var NodePoolSystemSharedPathLabelKey = fmt.Sprintf("shared.node-pool.%v/path", SchemeGroupVersion.String())

// WorkloadNodePoolAnnotationKey is the key that should be used in an annotation by
// workloads that run on a node pool.
var WorkloadNodePoolAnnotationKey = fmt.Sprintf("workload.%v/node-pools", SchemeGroupVersion.String())

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NodePoolSpec   `json:"spec"`
	Status            NodePoolStatus `json:"status,omitempty"`
}

func (np *NodePool) ID(epoch NodePoolEpoch) string {
	return fmt.Sprintf("%v.%v.%v", np.Namespace, np.Name, epoch)
}

type NodePoolType string

const (
	NodePoolTypeServiceDedicated NodePoolType = "service-dedicated"
	NodePoolTypeSystemShared     NodePoolType = "system-shared"
	NodePoolTypeUnknown          NodePoolType = "unknown"
)

func (np *NodePool) Type() NodePoolType {
	if np.Labels == nil {
		return NodePoolTypeUnknown
	}

	if _, ok := np.Labels[NodePoolServiceDedicatedIDLabelKey]; ok {
		return NodePoolTypeServiceDedicated
	}

	if _, ok := np.Labels[NodePoolSystemSharedPathLabelKey]; ok {
		return NodePoolTypeSystemShared
	}

	return NodePoolTypeUnknown
}

func (np *NodePool) TypeDescription() string {
	if np.Labels == nil {
		return "UNKNOWN"
	}

	if serviceID, ok := np.Labels[NodePoolServiceDedicatedIDLabelKey]; ok {
		return fmt.Sprintf("dedicated for service %v", serviceID)
	}

	if path, ok := np.Labels[NodePoolSystemSharedPathLabelKey]; ok {
		return fmt.Sprintf("shared node pool %v", path)
	}

	return "UNKNOWN"
}

func (np *NodePool) Description(namespacePrefix string) string {
	// TODO: when adding lattice node pools may have to adjust his
	systemID, err := kubeutil.SystemID(namespacePrefix, np.Namespace)
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

type NodePoolEpoch int64

func (np *NodePool) Affinity(epoch NodePoolEpoch) *corev1.NodeAffinity {
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

func (np *NodePool) Toleration(epoch NodePoolEpoch) corev1.Toleration {
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
	ObservedGeneration int64         `json:"observedGeneration"`
	State              NodePoolState `json:"state"`

	// Epochs is a mapping from an epoch to the status of that epoch.
	// An epoch is a manifestation of the node pool that requires replacing infrastructure.
	// For example, changing the instance type of a node pool will require new nodes,
	// and thus requires a new epoch of the node pool.
	// Changing the number of nodes in a node pool does not require replacing the
	// existing nodes, simply scaling them, and thus does not require a new epoch.
	Epochs NodePoolStatusEpochs `json:"epochs"`
}

type NodePoolStatusEpochs map[NodePoolEpoch]NodePoolStatusEpoch

func (e NodePoolStatusEpochs) Epochs() []NodePoolEpoch {
	var epochs []NodePoolEpoch
	for epoch := range e {
		epochs = append(epochs, epoch)
	}

	sort.Slice(epochs, func(i, j int) bool {
		return epochs[i] < epochs[j]
	})

	return epochs
}

func (e NodePoolStatusEpochs) CurrentEpoch() (NodePoolEpoch, bool) {
	epochs := e.Epochs()

	if len(epochs) == 0 {
		return 0, false
	}

	return epochs[len(epochs)-1], true
}

func (e NodePoolStatusEpochs) NextEpoch() NodePoolEpoch {
	current, ok := e.CurrentEpoch()
	if !ok {
		return 1
	}

	return current + 1
}

func (e NodePoolStatusEpochs) Epoch(epoch NodePoolEpoch) (*NodePoolStatusEpoch, bool) {
	status, ok := e[epoch]
	if !ok {
		return nil, false
	}

	return &status, true
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
	NumInstances int32         `json:"numInstances"`
	InstanceType string        `json:"instanceType"`
	State        NodePoolState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NodePool `json:"items"`
}

// NodePoolAnnotation is the type that should be the value of the WorkloadNodePoolAnnotationKey annotation.
// It maps from namespaces to a map of names to a slice of epochs.
// For example, if a service is currently running on the abc node pool in namespace bar, and is
// on both epochs 1 and 2 of that node pool, and is also currently running on the xyz node pool
// in namespace foo, and is running on epoch 5 of that node pool, the annotation should have
// the following value:
//	{
//		"bar": { "abc": [1, 2] },
//		"foo": { "xyz": [5] }
//	}
type NodePoolAnnotationValue map[string]map[string][]NodePoolEpoch

func (a NodePoolAnnotationValue) IsEmpty() bool {
	return len(a) == 0
}

func (a NodePoolAnnotationValue) NodePools(namespace string) (map[string][]NodePoolEpoch, bool) {
	nodePools, ok := a[namespace]
	return nodePools, ok
}

func (a NodePoolAnnotationValue) Add(namespace, nodePool string, epoch NodePoolEpoch) {
	nodePools, ok := a.NodePools(namespace)
	if !ok {
		nodePools = make(map[string][]NodePoolEpoch)
	}

	epochs, ok := nodePools[nodePool]
	if !ok {
		epochs = make([]NodePoolEpoch, 0)
	}

	containsEpoch := false
	for _, e := range epochs {
		if e == epoch {
			containsEpoch = true
			break
		}
	}

	if !containsEpoch {
		epochs = append(epochs, epoch)
	}

	sort.Slice(epochs, func(i, j int) bool {
		return epochs[i] < epochs[j]
	})

	nodePools[nodePool] = epochs
	a[namespace] = nodePools
}

func (a NodePoolAnnotationValue) Epochs(namespace, nodePool string) ([]NodePoolEpoch, bool) {
	nodePools, ok := a.NodePools(namespace)
	if !ok {
		return nil, false
	}

	epochs, ok := nodePools[nodePool]
	return epochs, ok
}

func (a NodePoolAnnotationValue) ContainsNodePool(namespace, nodePool string) bool {
	_, ok := a.Epochs(namespace, nodePool)
	return ok
}

func (a NodePoolAnnotationValue) ContainsLargerEpoch(namespace, nodePool string, epoch NodePoolEpoch) bool {
	epochs, ok := a.Epochs(namespace, nodePool)
	if !ok {
		return false
	}

	for _, e := range epochs {
		if e > epoch {
			return true
		}
	}

	return false
}

func (a NodePoolAnnotationValue) ContainsEpoch(namespace, nodePool string, epoch NodePoolEpoch) bool {
	epochs, ok := a.Epochs(namespace, nodePool)
	if !ok {
		return false
	}

	for _, e := range epochs {
		if e == epoch {
			return true
		}
	}

	return false
}
