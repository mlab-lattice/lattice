package backend

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListSystems() ([]types.System, error) {
	selector := labels.NewSelector()

	requirement, err := labels.NewRequirement(
		kubeconstants.LabelKeyLatticeClusterID,
		selection.Equals,
		[]string{string(kb.ClusterID)},
	)
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	listOptions := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	systems, err := kb.LatticeClient.LatticeV1().Systems(corev1.NamespaceAll).List(listOptions)
	if err != nil {
		return nil, err
	}

	var externalSystems []types.System
	for _, system := range systems.Items {
		externalSystem, err := kb.transformSystem(&system)
		if err != nil {
			return nil, err
		}

		externalSystems = append(externalSystems, *externalSystem)
	}

	return externalSystems, nil
}

func (kb *KubernetesBackend) GetSystem(id types.SystemID) (*types.System, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	system, err := kb.LatticeClient.LatticeV1().Systems(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	externalSystem, err := kb.transformSystem(system)
	return externalSystem, true, err
}

func (kb *KubernetesBackend) transformSystem(system *crv1.System) (*types.System, error) {
	name, err := kubeutil.SystemID(system.Namespace)
	if err != nil {
		return nil, err
	}

	externalSystem := &types.System{
		ID:            types.SystemID(name),
		State:         getSystemState(system.Status.State),
		DefinitionURL: system.Spec.DefinitionURL,
	}

	externalServices := map[tree.NodePath]types.Service{}
	for path, serviceName := range system.Status.Services {
		serviceStatus, ok := system.Status.ServiceStatuses[serviceName]
		if !ok {
			err := fmt.Errorf(
				"System %v has Service %v for %v but does not have its status",
				system.Namespace,
				serviceName,
				path,
			)
			return nil, err
		}

		externalService := kb.transformService(serviceName, path, &serviceStatus)
		externalServices[path] = externalService
	}

	externalSystem.Services = externalServices
	return externalSystem, nil
}

func getSystemState(state crv1.SystemState) types.SystemState {
	switch state {
	case crv1.SystemStateScaling:
		return types.SystemStateScaling
	case crv1.SystemStateUpdating:
		return types.SystemStateUpdating
	case crv1.SystemStateStable:
		return types.SystemStateStable
	case crv1.SystemStateFailed:
		return types.SystemStateFailed
	default:
		panic("unreachable")
	}
}
