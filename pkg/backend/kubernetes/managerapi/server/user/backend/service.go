package backend

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListServices(id types.SystemID) ([]types.Service, error) {
	services, err := kb.LatticeClient.LatticeV1().Services(string(id)).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []types.Service
	for _, service := range services.Items {
		externalService := kb.transformService(service.Name, service.Spec.Path, &service.Status)
		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (kb *KubernetesBackend) GetService(id types.SystemID, path tree.NodePath) (*types.Service, error) {
	config, err := kb.getConfig()
	if err != nil {
		return nil, err
	}

	selector := kubelabels.NewSelector()
	requirement, err := kubelabels.NewRequirement(kubeconstants.LabelKeyServicePath, selection.Equals, []string{string(path)})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	listOptions := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	namespace := kubeutil.SystemNamespace(string(id), config.Spec.KubernetesNamespacePrefix)
	services, err := kb.LatticeClient.LatticeV1().Services(namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	if len(services.Items) > 1 {
		return nil, fmt.Errorf("found multiple Services for System %v %v", id, path)
	}

	if len(services.Items) == 0 {
		return nil, nil
	}

	service := services.Items[0]
	externalService := kb.transformService(service.Name, service.Spec.Path, &service.Status)
	return &externalService, nil
}

func (kb *KubernetesBackend) transformService(serviceName string, path tree.NodePath, serviceStatus *crv1.ServiceStatus) types.Service {
	return types.Service{
		ID:               types.ServiceID(serviceName),
		Path:             path,
		State:            getServicedState(serviceStatus.State),
		UpdatedInstances: serviceStatus.UpdatedInstances,
		StaleInstances:   serviceStatus.StaleInstances,
		// FIXME: add GetServicePublicAddress to provider
	}
}

func getServicedState(state crv1.ServiceState) types.ServiceState {
	switch state {
	case crv1.ServiceStatePending:
		return types.ServiceStatePending
	case crv1.ServiceStateScalingDown:
		return types.ServiceStateScalingDown
	case crv1.ServiceStateScalingUp:
		return types.ServiceStateScalingUp
	case crv1.ServiceStateUpdating:
		return types.ServiceStateUpdating
	case crv1.ServiceStateStable:
		return types.ServiceStateStable
	case crv1.ServiceStateFailed:
		return types.ServiceStateFailed
	default:
		panic("unreachable")
	}
}
