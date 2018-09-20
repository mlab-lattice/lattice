package system

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type nodePoolBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *nodePoolBackend) List() ([]v1.NodePool, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	nodePools, err := b.backend.latticeClient.LatticeV1().NodePools(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalNodePools []v1.NodePool
	for _, nodePool := range nodePools.Items {
		path, err := b.getNodePoolPath(&nodePool)
		if err != nil {
			return nil, err
		}

		externalNodePool, err := b.transformNodePool(nodePool.Name, path, &nodePool.Status)
		if err != nil {
			return nil, err
		}

		externalNodePools = append(externalNodePools, externalNodePool)
	}

	return externalNodePools, nil
}

func (b *nodePoolBackend) Get(path tree.PathSubcomponent) (*v1.NodePool, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	var selector labels.Selector
	if path.Subcomponent() == "" {
		selector = labels.NewSelector()
		requirement, err := labels.NewRequirement(
			latticev1.ServicePathLabelKey,
			selection.Equals,
			[]string{path.Path().ToDomain()},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
		services, err := b.backend.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}

		if len(services.Items) > 1 {
			return nil, fmt.Errorf("found multiple services for path %v", path.Path().String())
		}

		if len(services.Items) == 0 {
			return nil, nil
		}

		service := services.Items[0]

		selector = labels.NewSelector()
		requirement, err = labels.NewRequirement(
			latticev1.NodePoolServiceDedicatedIDLabelKey,
			selection.Equals,
			[]string{service.Name},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
	} else {
		requirement, err := labels.NewRequirement(
			latticev1.NodePoolSystemSharedPathLabelKey,
			selection.Equals,
			[]string{path.Path().ToDomain()},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)

		requirement, err = labels.NewRequirement(
			latticev1.NodePoolSystemSharedNameLabelKey,
			selection.Equals,
			[]string{path.Subcomponent()},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
	}

	nodePools, err := b.backend.latticeClient.LatticeV1().NodePools(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(nodePools.Items) > 1 {
		return nil, fmt.Errorf("found multiple node pools for selector %v", selector.String())
	}

	if len(nodePools.Items) == 0 {
		return nil, nil
	}

	nodePool := &nodePools.Items[0]
	externalNodePool, err := b.transformNodePool(nodePool.Name, path, &nodePool.Status)
	if err != nil {
		return nil, err
	}

	return &externalNodePool, nil
}

func (b *nodePoolBackend) getNodePoolPath(nodePool *latticev1.NodePool) (tree.PathSubcomponent, error) {
	serviceID, ok := nodePool.ServiceDedicatedIDLabel()
	if ok {
		service, err := b.backend.latticeClient.LatticeV1().Services(nodePool.Namespace).Get(serviceID, metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}

			return "", nil
		}

		servicePath, err := service.PathLabel()
		if err != nil {
			return "", err
		}

		path, err := tree.NewPathSubcomponentFromParts(servicePath, "node_pool")
		if err != nil {
			return "", err
		}

		return path, nil
	}

	path, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil {
		return "", err
	}

	if ok {
		return path, nil
	}

	return "", fmt.Errorf("%v did not contain service id or system shared path labels", nodePool.Description(b.backend.namespacePrefix))
}

func (b *nodePoolBackend) transformNodePool(id string, path tree.PathSubcomponent, status *latticev1.NodePoolStatus) (v1.NodePool, error) {
	state, err := getNodePoolState(status.State)
	if err != nil {
		return v1.NodePool{}, err
	}

	var failureInfo *v1.NodePoolFailureInfo
	if status.FailureInfo != nil {
		failureInfo = &v1.NodePoolFailureInfo{
			Time:    status.FailureInfo.Timestamp.Time,
			Message: status.FailureInfo.Message,
		}
	}

	instanceType := "unknown"
	currentEpoch, ok := status.Epochs.CurrentEpoch()
	if ok {
		epoch, ok := status.Epochs.Epoch(currentEpoch)
		if !ok {
			return v1.NodePool{}, fmt.Errorf("node pool %v had current epoch %v but does not have its status", id, currentEpoch)
		}

		instanceType = epoch.Spec.InstanceType
	}

	var numInstances int32 = 0
	for _, epoch := range status.Epochs {
		numInstances += epoch.Status.NumInstances
	}

	nodePool := v1.NodePool{
		ID:   id,
		Path: path,

		State:       state,
		FailureInfo: failureInfo,

		InstanceType: instanceType,
		NumInstances: numInstances,
	}

	return nodePool, nil
}

func getNodePoolState(state latticev1.NodePoolState) (v1.NodePoolState, error) {
	switch state {
	case latticev1.NodePoolStatePending:
		return v1.NodePoolStatePending, nil
	case latticev1.NodePoolStateDeleting:
		return v1.NodePoolStateDeleting, nil

	case latticev1.NodePoolStateScaling:
		return v1.NodePoolStateScaling, nil
	case latticev1.NodePoolStateUpdating:
		return v1.NodePoolStateUpdating, nil
	case latticev1.NodePoolStateStable:
		return v1.NodePoolStateStable, nil
	case latticev1.NodePoolStateFailed:
		return v1.NodePoolStateFailed, nil
	default:
		return "", fmt.Errorf("invalid node pool state: %v", state)
	}
}
