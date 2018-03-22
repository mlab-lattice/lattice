package backend

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) CreateSystem(id types.SystemID, definitionURL string) (*types.System, error) {
	system := &latticev1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name:       string(id),
			Finalizers: []string{kubeconstants.KubeFinalizerSystemController},
		},
		Spec: latticev1.SystemSpec{
			DefinitionURL: definitionURL,
		},
		Status: latticev1.SystemStatus{
			State: latticev1.SystemStatePending,
		},
	}

	system, err := kb.latticeClient.LatticeV1().Systems(kubeutil.InternalNamespace(kb.latticeID)).Create(system)
	if err != nil {
		return nil, err
	}

	return kb.transformSystem(system)
}

func (kb *KubernetesBackend) ListSystems() ([]types.System, error) {
	listOptions := metav1.ListOptions{}
	systems, err := kb.latticeClient.LatticeV1().Systems(kubeutil.InternalNamespace(kb.latticeID)).List(listOptions)
	if err != nil {
		return nil, err
	}

	externalSystems := make([]types.System, 0)
	for _, system := range systems.Items {
		externalSystem, err := kb.transformSystem(&system)
		if err != nil {
			return nil, err
		}

		externalSystems = append(externalSystems, *externalSystem)
	}

	return externalSystems, nil
}

func (kb *KubernetesBackend) GetSystem(systemID types.SystemID) (*types.System, bool, error) {
	system, err := kb.latticeClient.LatticeV1().Systems(kubeutil.InternalNamespace(kb.latticeID)).Get(string(systemID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	externalSystem, err := kb.transformSystem(system)
	return externalSystem, true, err
}

func (kb *KubernetesBackend) DeleteSystem(systemID types.SystemID) error {
	return kb.latticeClient.LatticeV1().Systems(kubeutil.InternalNamespace(kb.latticeID)).Delete(string(systemID), nil)
}

func (kb *KubernetesBackend) transformSystem(system *latticev1.System) (*types.System, error) {
	var state types.SystemState
	if system.DeletionTimestamp != nil {
		state = types.SystemStateDeleting
	} else {
		state = getSystemState(system.Status.State)
	}

	externalSystem := &types.System{
		ID:            types.SystemID(system.Name),
		State:         state,
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

func getSystemState(state latticev1.SystemState) types.SystemState {
	switch state {
	case latticev1.SystemStatePending:
		return types.SystemStatePending
	case latticev1.SystemStateFailed:
		return types.SystemStateFailed

	case latticev1.SystemStateStable:
		return types.SystemStateStable
	case latticev1.SystemStateDegraded:
		return types.SystemStateDegraded
	case latticev1.SystemStateScaling:
		return types.SystemStateScaling
	case latticev1.SystemStateUpdating:
		return types.SystemStateUpdating
	default:
		panic("unreachable")
	}
}
