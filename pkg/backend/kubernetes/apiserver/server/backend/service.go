package backend

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListServices(systemID v1.SystemID) ([]v1.Service, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)

	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []v1.Service
	for _, service := range services.Items {
		externalService := kb.transformService(service.Name, service.Spec.Path, &service.Status)
		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (kb *KubernetesBackend) GetService(id v1.SystemID, path tree.NodePath) (*v1.Service, error) {
	selector := kubelabels.NewSelector()
	requirement, err := kubelabels.NewRequirement(
		kubeconstants.LabelKeyServicePathDomain,
		selection.Equals,
		[]string{path.ToDomain(true)},
	)
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	listOptions := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(listOptions)
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

func (kb *KubernetesBackend) transformService(
	serviceName string,
	path tree.NodePath,
	serviceStatus *latticev1.ServiceStatus,
) v1.Service {
	service := v1.Service{
		ID:               v1.ServiceID(serviceName),
		Path:             path,
		State:            getServicedState(serviceStatus.State),
		UpdatedInstances: serviceStatus.UpdatedInstances,
		StaleInstances:   serviceStatus.StaleInstances,
	}

	ports := v1.ServicePublicPorts{}
	for port, portInfo := range serviceStatus.PublicPorts {
		ports[port] = v1.ServicePublicPort{
			Address: portInfo.Address,
		}
	}
	service.PublicPorts = ports

	var failureMessage *string
	if serviceStatus.FailureInfo != nil {
		internalError := "internal error"
		failureMessage = &internalError

		if !serviceStatus.FailureInfo.Internal {
			errorMessage := fmt.Sprintf("%v: %v", serviceStatus.FailureInfo.Time, serviceStatus.FailureInfo.Message)
			failureMessage = &errorMessage
		}
	}
	service.FailureMessage = failureMessage

	return service
}

func getServicedState(state latticev1.ServiceState) v1.ServiceState {
	switch state {
	case latticev1.ServiceStatePending:
		return v1.ServiceStatePending
	case latticev1.ServiceStateScalingDown:
		return v1.ServiceStateScalingDown
	case latticev1.ServiceStateScalingUp:
		return v1.ServiceStateScalingUp
	case latticev1.ServiceStateUpdating:
		return v1.ServiceStateUpdating
	case latticev1.ServiceStateStable:
		return v1.ServiceStateStable
	case latticev1.ServiceStateFailed:
		return v1.ServiceStateFailed
	default:
		panic("unreachable")
	}
}
