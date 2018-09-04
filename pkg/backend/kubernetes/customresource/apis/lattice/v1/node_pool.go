package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strconv"
	"strings"
)

const (
	ResourceSingularNodePool = "nodepool"
	ResourcePluralNodePool   = "nodepools"
	ResourceScopeNodePool    = apiextensionsv1beta1.NamespaceScoped
)

var (
	NodePoolKind     = SchemeGroupVersion.WithKind("NodePool")
	NodePoolListKind = SchemeGroupVersion.WithKind("NodePoolList")

	NodePoolIDLabelKey = fmt.Sprintf("node-pool.%v/id", GroupName)

	// NodePoolServiceDedicatedID is the key for a label indicating that the node pool is dedicated for a service.
	// The label's value should be the ID of the service.
	// TODO: should we just use ServiceIDLabelKey here instead? if so what do we use for shared/lattice node pools
	NodePoolServiceDedicatedIDLabelKey = fmt.Sprintf("service.dedicated.node-pool.%v/id", GroupName)

	// NodePoolSystemSharedPathLabelKey is the key for a label indicating that the node pool is shared for a system.
	// The label's value should be the node pool's path in the system definition.
	NodePoolSystemSharedPathLabelKey = fmt.Sprintf("shared.node-pool.%v/path", GroupName)
	NodePoolSystemSharedNameLabelKey = fmt.Sprintf("shared.node-pool.%v/name", GroupName)

	// NodePoolWorkloadAnnotationKey is the key that should be used in an annotation by
	// workloads that run on a node pool.
	NodePoolWorkloadAnnotationKey = fmt.Sprintf("workload.%v/node-pools", GroupName)

	AllNodePoolsSelector = corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      NodePoolIDLabelKey,
						Operator: corev1.NodeSelectorOpExists,
					},
				},
			},
		},
	}

	AllNodePoolAffinity = corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &AllNodePoolsSelector,
	}

	AllNodePoolTolleration = corev1.Toleration{
		Key:      NodePoolIDLabelKey,
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}
)

func NodePoolIDLabelInfo(namespacePrefix, value string) (v1.SystemID, string, NodePoolEpoch, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return "", "", 0, fmt.Errorf("malformed node pool ID label")
	}

	systemID, err := kubeutil.SystemID(namespacePrefix, parts[0])
	if err != nil {
		return "", "", 0, err
	}

	nodePoolID := parts[1]
	epochStr := parts[2]

	epoch, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("error converting epoch string value %v: %v", epochStr, err)
	}

	return systemID, nodePoolID, NodePoolEpoch(epoch), nil
}

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
)

func (np *NodePool) ServiceDedicatedIDLabel() (string, bool) {
	serviceID, ok := np.Labels[NodePoolServiceDedicatedIDLabelKey]
	return serviceID, ok
}

func (np *NodePool) SystemSharedPathLabel() (tree.PathSubcomponent, bool, error) {
	pathLabel, ok := np.Labels[NodePoolSystemSharedPathLabelKey]
	if !ok {
		return "", false, nil
	}

	nameLabel, ok := np.Labels[NodePoolSystemSharedNameLabelKey]
	if !ok {
		return "", false, nil
	}

	path, err := tree.NewPathFromDomain(pathLabel)
	if err != nil {
		return "", false, err
	}

	subcomponent, err := tree.NewPathSubcomponentFromParts(path, nameLabel)
	if err != nil {
		return "", false, err
	}

	return subcomponent, true, nil
}

func (np *NodePool) TypeDescription() string {
	if np.Labels == nil {
		return "UNKNOWN"
	}

	if serviceID, ok := np.ServiceDedicatedIDLabel(); ok {
		return fmt.Sprintf("dedicated for service %v", serviceID)
	}

	if path, ok, err := np.SystemSharedPathLabel(); err == nil && ok {
		return fmt.Sprintf("shared node pool %v", path)
	}

	return "UNKNOWN"
}

func (np *NodePool) Description(namespacePrefix string) string {
	// TODO: when adding lattice node pools may have to adjust this
	systemID, err := kubeutil.SystemID(namespacePrefix, np.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", np.Namespace))
	}

	return fmt.Sprintf("node pool %v (%v in system %v)", np.Name, np.TypeDescription(), systemID)
}

func (np *NodePool) Stable() bool {
	return np.UpdateProcessed() && np.Status.State == NodePoolStateStable
}

func (np *NodePool) Failed() bool {
	return np.UpdateProcessed() && np.Status.State == NodePoolStateFailed
}

func (np *NodePool) UpdateProcessed() bool {
	return np.Status.ObservedGeneration >= np.Generation
}

func (np *NodePool) Reason() string {
	if !np.UpdateProcessed() {
		return "waiting for update to be processed"
	}

	switch np.Status.State {
	case NodePoolStateStable:
		return ""
	case NodePoolStateUpdating:
		return "updating"
	case NodePoolStatePending:
		return "pending"
	case NodePoolStateScaling:
		return "scaling"
	case NodePoolStateDeleting:
		return "deleting"
	case NodePoolStateFailed:
		failureReason := "unknown reason"
		if np.Status.FailureInfo != nil {
			failureReason = fmt.Sprintf(
				"%v at %v",
				np.Status.FailureInfo.Message,
				np.Status.FailureInfo.Timestamp.String(),
			)
		}

		return fmt.Sprintf("failed: %v", failureReason)
	default:
		return fmt.Sprintf("in unknown state: %v", np.Status.State)
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodePoolEpoch int64

func (np *NodePool) Affinity(epoch NodePoolEpoch) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      NodePoolIDLabelKey,
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
		Key:      NodePoolIDLabelKey,
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
	ObservedGeneration int64 `json:"observedGeneration"`

	State       NodePoolState              `json:"state"`
	FailureInfo *NodePoolStatusFailureInfo `json:"failureInfo"`

	// Epochs is a mapping from an epoch to the status of that epoch.
	// An epoch is a manifestation of the node pool that requires replacing infrastructure.
	// For example, changing the instance type of a node pool will require new nodes,
	// and thus requires a new epoch of the node pool.
	// Changing the number of nodes in a node pool does not require replacing the
	// existing nodes, simply scaling them, and thus does not require a new epoch.
	Epochs NodePoolStatusEpochs `json:"epochs"`
}

type NodePoolStatusFailureInfo struct {
	Message   string      `json:"message"`
	Timestamp metav1.Time `json:"time"`
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
	NodePoolStatePending  NodePoolState = ""
	NodePoolStateScaling  NodePoolState = "scaling"
	NodePoolStateUpdating NodePoolState = "updating"
	NodePoolStateStable   NodePoolState = "stable"
	NodePoolStateFailed   NodePoolState = "failed"
	NodePoolStateDeleting NodePoolState = "deleting"
)

type NodePoolStatusEpoch struct {
	Spec   NodePoolSpec              `json:"spec"`
	Status NodePoolStatusEpochStatus `json:"status"`
}

type NodePoolStatusEpochStatus struct {
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

// NodePoolAnnotation is the type that should be the value of the NodePoolWorkloadAnnotationKey annotation.
// It maps from namespaces to a map of names to a slice of epochs.
// For example, if a service is currently running on the abc node pool in namespace bar, and is
// on both epochs 1 and 2 of that node pool, and is also currently running on the xyz node pool
// in namespace foo, and is running on epoch 5 of that node pool, the annotation should have
// the following value:
//	{
//		"bar": { "abc": [1, 2] },
//		"foo": { "xyz": [5] }
//	}
// +k8s:deepcopy-gen=false
type NodePoolAnnotationValue map[string]NodePoolAnnotationValueNamespace

// +k8s:deepcopy-gen=false
type NodePoolAnnotationValueNamespace map[string][]NodePoolEpoch

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
