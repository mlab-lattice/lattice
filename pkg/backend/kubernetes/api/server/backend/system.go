package backend

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) CreateSystem(id v1.SystemID, definitionURL string) (*v1.System, error) {
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

	system, err := kb.latticeClient.LatticeV1().Systems(kb.internalNamespace()).Create(system)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil, v1.NewSystemAlreadyExistsError(id)
		}

		return nil, err
	}

	return kb.transformSystem(system)
}

func (kb *KubernetesBackend) ListSystems() ([]v1.System, error) {
	listOptions := metav1.ListOptions{}
	systems, err := kb.latticeClient.LatticeV1().Systems(kb.internalNamespace()).List(listOptions)
	if err != nil {
		return nil, err
	}

	externalSystems := make([]v1.System, 0)
	for _, system := range systems.Items {
		externalSystem, err := kb.transformSystem(&system)
		if err != nil {
			return nil, err
		}

		externalSystems = append(externalSystems, *externalSystem)
	}

	return externalSystems, nil
}

func (kb *KubernetesBackend) GetSystem(systemID v1.SystemID) (*v1.System, error) {
	system, err := kb.latticeClient.LatticeV1().Systems(kb.internalNamespace()).Get(string(systemID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidSystemIDError(systemID)
		}

		return nil, err
	}

	return kb.transformSystem(system)
}

func (kb *KubernetesBackend) DeleteSystem(systemID v1.SystemID) error {
	err := kb.latticeClient.LatticeV1().Systems(kb.internalNamespace()).Delete(string(systemID), nil)
	if err == nil {
		return nil
	}

	if errors.IsConflict(err) {
		return v1.NewConflictError("")
	}

	return err
}

func (kb *KubernetesBackend) transformSystem(system *latticev1.System) (*v1.System, error) {
	var state v1.SystemState
	if system.DeletionTimestamp != nil {
		state = v1.SystemStateDeleting
	} else {
		var err error
		state, err = getSystemState(system.Status.State, system.UpdateProcessed())
		if err != nil {
			return nil, err
		}
	}

	externalSystem := &v1.System{
		ID:            v1.SystemID(system.Name),
		State:         state,
		DefinitionURL: system.Spec.DefinitionURL,
	}

	externalServices := map[tree.NodePath]v1.Service{}
	for path, status := range system.Status.Services {
		externalService, err := kb.transformService(path, &status.ServiceStatus)
		if err != nil {
			return nil, err
		}

		externalServices[path] = externalService
	}

	externalSystem.Services = externalServices
	return externalSystem, nil
}

func getSystemState(state latticev1.SystemState, updateProcessed bool) (v1.SystemState, error) {
	// If the system is pending or failed, it doesn't matter if the controller has seen the most
	// recent spec
	if state == latticev1.SystemStatePending {
		return v1.SystemStatePending, nil
	}

	if state == latticev1.SystemStateFailed {
		return v1.SystemStateFailed, nil
	}

	// If the system is in a created state, but the controller has not yet seen the most up to date
	// spec, then the system is updating
	if !updateProcessed {
		return v1.SystemStateUpdating, nil
	}

	// If the controller has seen the most recent spec, then we can return the true system status
	switch state {
	case latticev1.SystemStateStable:
		return v1.SystemStateStable, nil
	case latticev1.SystemStateDegraded:
		return v1.SystemStateDegraded, nil
	case latticev1.SystemStateScaling:
		return v1.SystemStateScaling, nil
	case latticev1.SystemStateUpdating:
		return v1.SystemStateUpdating, nil
	default:
		return "", fmt.Errorf("invalid system state: %v", state)
	}
}

func (kb *KubernetesBackend) ensureSystemCreated(systemID v1.SystemID) (*v1.System, error) {
	system, err := kb.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	switch system.State {
	case v1.SystemStatePending, v1.SystemStateFailed, v1.SystemStateDeleting:
		return system, v1.NewSystemNotCreatedError(systemID, system.State)
	case v1.SystemStateStable, v1.SystemStateDegraded, v1.SystemStateScaling, v1.SystemStateUpdating:
		return system, nil
	default:
		return nil, fmt.Errorf("invalid system state: %v", system.State)
	}
}
