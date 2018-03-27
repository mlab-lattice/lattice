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
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)

	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []v1.Service
	for _, service := range services.Items {
		externalService, err := kb.transformService(v1.ServiceID(service.Name), service.Spec.Path, &service.Status)
		if err != nil {
			return nil, err
		}

		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (kb *KubernetesBackend) GetService(systemID v1.SystemID, path tree.NodePath) (*v1.Service, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

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

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	if len(services.Items) > 1 {
		return nil, fmt.Errorf("found multiple Services for System %v %v", systemID, path)
	}

	if len(services.Items) == 0 {
		return nil, nil
	}

	service := services.Items[0]
	externalService, err := kb.transformService(v1.ServiceID(service.Name), service.Spec.Path, &service.Status)
	if err != nil {
		return nil, err
	}

	return &externalService, nil
}

func (kb *KubernetesBackend) transformService(id v1.ServiceID, path tree.NodePath, status *latticev1.ServiceStatus) (v1.Service, error) {
	state, err := getServiceState(status.State)
	if err != nil {
		return v1.Service{}, err
	}

	service := v1.Service{
		ID:               id,
		Path:             path,
		State:            state,
		UpdatedInstances: status.UpdatedInstances,
		StaleInstances:   status.StaleInstances,
	}

	ports := v1.ServicePublicPorts{}
	for port, portInfo := range status.PublicPorts {
		ports[port] = v1.ServicePublicPort{
			Address: portInfo.Address,
		}
	}
	service.PublicPorts = ports

	var failureMessage *string
	if status.FailureInfo != nil {
		internalError := "internal error"
		failureMessage = &internalError

		if !status.FailureInfo.Internal {
			errorMessage := fmt.Sprintf("%v: %v", status.FailureInfo.Time, status.FailureInfo.Message)
			failureMessage = &errorMessage
		}
	}
	service.FailureMessage = failureMessage

	return service, nil
}

func getServiceState(state latticev1.ServiceState) (v1.ServiceState, error) {
	switch state {
	case latticev1.ServiceStatePending:
		return v1.ServiceStatePending, nil
	case latticev1.ServiceStateScalingDown:
		return v1.ServiceStateScalingDown, nil
	case latticev1.ServiceStateScalingUp:
		return v1.ServiceStateScalingUp, nil
	case latticev1.ServiceStateUpdating:
		return v1.ServiceStateUpdating, nil
	case latticev1.ServiceStateStable:
		return v1.ServiceStateStable, nil
	case latticev1.ServiceStateFailed:
		return v1.ServiceStateFailed, nil
	default:
		return "", fmt.Errorf("invalid service state: %v", state)
	}
}
