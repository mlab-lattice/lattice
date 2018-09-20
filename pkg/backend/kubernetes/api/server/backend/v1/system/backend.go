package system

import (
	"fmt"

	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeclientset "k8s.io/client-go/kubernetes"
)

func NewBackend(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) *Backend {
	return &Backend{namespacePrefix, kubeClient, latticeClient}
}

type Backend struct {
	namespacePrefix string
	kubeClient      kubeclientset.Interface
	latticeClient   latticeclientset.Interface
}

func (b *Backend) Create(id v1.SystemID, definitionURL string) (*v1.System, error) {
	system := &latticev1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(id),
		},
		Spec: latticev1.SystemSpec{
			DefinitionURL: definitionURL,
		},
	}

	system, err := b.latticeClient.LatticeV1().Systems(b.internalNamespace()).Create(system)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil, v1.NewSystemAlreadyExistsError()
		}

		return nil, err
	}

	return b.transformSystem(system)
}

func (b *Backend) List() ([]v1.System, error) {
	listOptions := metav1.ListOptions{}
	systems, err := b.latticeClient.LatticeV1().Systems(b.internalNamespace()).List(listOptions)
	if err != nil {
		return nil, err
	}

	externalSystems := make([]v1.System, 0)
	for _, system := range systems.Items {
		externalSystem, err := b.transformSystem(&system)
		if err != nil {
			return nil, err
		}

		externalSystems = append(externalSystems, *externalSystem)
	}

	return externalSystems, nil
}

func (b *Backend) Get(id v1.SystemID) (*v1.System, error) {
	system, err := b.latticeClient.LatticeV1().Systems(b.internalNamespace()).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidSystemIDError()
		}

		return nil, err
	}

	return b.transformSystem(system)
}

func (b *Backend) Delete(id v1.SystemID) error {
	err := b.latticeClient.LatticeV1().Systems(b.internalNamespace()).Delete(string(id), nil)
	if err == nil {
		return nil
	}

	if errors.IsConflict(err) {
		return v1.NewConflictError()
	}

	if errors.IsNotFound(err) {
		return v1.NewInvalidSystemIDError()
	}

	return err
}

func (b *Backend) transformSystem(system *latticev1.System) (*v1.System, error) {
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

	var version *v1.Version
	versionLabel, ok := system.DefinitionVersionLabel()
	if !ok {
		version = &versionLabel
	}

	externalSystem := &v1.System{
		ID:            v1.SystemID(system.Name),
		DefinitionURL: system.Spec.DefinitionURL,
		Status: v1.SystemStatus{
			State: state,

			Version: version,
		},
	}

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

func (b *Backend) ensureSystemCreated(id v1.SystemID) (*v1.System, error) {
	system, err := b.Get(id)
	if err != nil {
		return nil, err
	}

	switch system.Status.State {
	case v1.SystemStateDeleting:
		return system, v1.NewSystemDeletingError()
	case v1.SystemStateFailed:
		return system, v1.NewSystemFailedError()
	case v1.SystemStatePending:
		return system, v1.NewSystemPendingError()
	case v1.SystemStateStable, v1.SystemStateDegraded, v1.SystemStateScaling, v1.SystemStateUpdating:
		return system, nil
	default:
		return nil, fmt.Errorf("invalid system state: %v", system.Status.State)
	}
}

func (b *Backend) systemNamespace(systemID v1.SystemID) string {
	return kubeutil.SystemNamespace(b.namespacePrefix, systemID)
}

func (b *Backend) internalNamespace() string {
	return kubeutil.InternalNamespace(b.namespacePrefix)
}

func (b *Backend) Builds(system v1.SystemID) serverv1.SystemBuildBackend {
	return &buildBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) Deploys(system v1.SystemID) serverv1.SystemDeployBackend {
	return &deployBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) Jobs(system v1.SystemID) serverv1.SystemJobBackend {
	return &jobBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) NodePools(system v1.SystemID) serverv1.SystemNodePoolBackend {
	return &nodePoolBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) Secrets(system v1.SystemID) serverv1.SystemSecretBackend {
	return &secretBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) Services(system v1.SystemID) serverv1.SystemServiceBackend {
	return &serviceBackend{
		backend: b,
		system:  system,
	}
}

func (b *Backend) Teardowns(system v1.SystemID) serverv1.SystemTeardownBackend {
	return &teardownBackend{
		backend: b,
		system:  system,
	}
}
