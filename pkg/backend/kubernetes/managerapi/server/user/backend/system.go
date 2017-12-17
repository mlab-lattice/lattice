package backend

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/lattice"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) GetSystem(id types.SystemID) (*types.System, error) {
	config, err := kb.getConfig()
	if err != nil {
		return nil, err
	}

	namespace := latticeutil.SystemNamespace(string(id), config.Spec.KubernetesNamespacePrefix)
	system, err := kb.LatticeClient.LatticeV1().Systems(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return kb.transformSystem(system)
}

func (kb *KubernetesBackend) transformSystem(system *crv1.System) (*types.System, error) {
	name, err := latticeutil.SystemName(system.Namespace)
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
