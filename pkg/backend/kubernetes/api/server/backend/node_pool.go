package backend

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListNodePools(systemID v1.SystemID) ([]v1.NodePool, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	nodePools, err := kb.latticeClient.LatticeV1().NodePools(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalNodePools []v1.NodePool
	for _, nodePool := range nodePools.Items {
		path, err := kb.getNodePoolPath(&nodePool)
		if err != nil {
			return nil, err
		}

		externalNodePool, err := kb.transformNodePool(nodePool.Name, path, &nodePool.Status)
		if err != nil {
			return nil, err
		}

		externalNodePools = append(externalNodePools, externalNodePool)
	}

	return externalNodePools, nil
}

func (kb *KubernetesBackend) GetNodePool(systemID v1.SystemID, path v1.NodePoolPath) (*v1.NodePool, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	var selector labels.Selector
	if path.Name == nil {
		selector = labels.NewSelector()
		requirement, err := labels.NewRequirement(
			latticev1.ServicePathLabelKey,
			selection.Equals,
			[]string{path.Path.ToDomain()},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
		services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}

		if len(services.Items) > 1 {
			return nil, fmt.Errorf("found multiple services for path %v", path.Path.String())
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
			[]string{path.Path.ToDomain()},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)

		requirement, err = labels.NewRequirement(
			latticev1.NodePoolSystemSharedNameLabelKey,
			selection.Equals,
			[]string{*path.Name},
		)
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
	}

	nodePools, err := kb.latticeClient.LatticeV1().NodePools(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
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
	externalNodePool, err := kb.transformNodePool(nodePool.Name, path, &nodePool.Status)
	if err != nil {
		return nil, err
	}

	return &externalNodePool, nil
}

func (kb *KubernetesBackend) getNodePoolPath(nodePool *latticev1.NodePool) (v1.NodePoolPath, error) {
	serviceID, ok := nodePool.ServiceDedicatedIDLabel()
	if ok {
		service, err := kb.latticeClient.LatticeV1().Services(nodePool.Namespace).Get(serviceID, metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				return v1.NodePoolPath{}, err
			}

			return v1.NodePoolPath{}, nil
		}

		servicePath, err := service.PathLabel()
		if err != nil {
			return v1.NodePoolPath{}, err
		}

		return v1.NewServiceNodePoolPath(servicePath), nil
	}

	path, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil {
		return v1.NodePoolPath{}, err
	}

	if ok {
		return path, nil
	}

	return v1.NodePoolPath{}, fmt.Errorf("%v did not contain service id or system shared path labels", nodePool.Description(kb.namespacePrefix))
}

func (kb *KubernetesBackend) transformNodePool(id string, path v1.NodePoolPath, status *latticev1.NodePoolStatus) (v1.NodePool, error) {
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
		Path: path.String(),

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
