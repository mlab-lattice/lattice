package backend

import (
	"fmt"
	"io"

	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListServices(systemID v1.SystemID) ([]v1.Service, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []v1.Service
	for _, service := range services.Items {
		servicePath, err := service.PathLabel()
		if err != nil {
			return nil, err
		}

		externalService, err := kb.transformService(service.Name, servicePath, &service.Status, namespace)
		if err != nil {
			return nil, err
		}

		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (kb *KubernetesBackend) GetService(systemID v1.SystemID, serviceID v1.ServiceID) (*v1.Service, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	service, err := kb.latticeClient.LatticeV1().Services(namespace).Get(string(serviceID), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	servicePath, err := service.PathLabel()

	if err != nil {
		return nil, err
	}

	externalService, err := kb.transformService(service.Name, servicePath, &service.Status, namespace)
	if err != nil {
		return nil, err
	}

	return &externalService, nil
}

func (kb *KubernetesBackend) GetServiceByPath(systemID v1.SystemID, path tree.NodePath) (*v1.Service, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	// TODO fixme try to push query to kube api instead of manually filtering it here
	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, service := range services.Items {

		servicePath, err := service.PathLabel()
		if err != nil {
			return nil, err
		}

		if path == servicePath {
			externalService, err := kb.transformService(service.Name, servicePath, &service.Status, namespace)
			if err != nil {
				return nil, err
			}

			return &externalService, nil
		}
	}

	return nil, nil
}
func (kb *KubernetesBackend) ServiceLogs(systemID v1.SystemID, serviceID v1.ServiceID, component string,
	instance string, follow bool) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)

	_, err := kb.latticeClient.LatticeV1().Services(namespace).Get(string(serviceID), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pod, err := kb.findServicePod(serviceID, instance, namespace)

	if err != nil {
		return nil, err
	}

	container := kubeutil.UserResourcePrefix + component
	logOptions := &corev1.PodLogOptions{Follow: follow, Container: container}

	req := kb.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, logOptions)
	return req.Stream()

}

// findServicePod finds service pod by instance id or service's single pod if id was not specified
func (kb *KubernetesBackend) findServicePod(serviceId v1.ServiceID, instance string, namespace string) (*corev1.Pod, error) {

	// check if instance was specified
	if instance != "" {
		podName := toServiceInstanceFullID(serviceId, instance)
		pod, err := kb.kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error fetching pod for instance %v/%v", namespace, podName)
		}
		return pod, nil
	}

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{string(serviceId)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v/%v job lookup: %v", namespace, serviceId, err)
	}

	selector = selector.Add(*requirement)
	pods, err := kb.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("found multiple pods for %v/%v. Need to specify an instance", namespace, serviceId)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods for %v/%v", namespace, serviceId)
	}

	return &pods.Items[0], nil

}

func (kb *KubernetesBackend) transformService(id string, path tree.NodePath, status *latticev1.ServiceStatus,
	namespace string) (v1.Service, error) {
	state, err := getServiceState(status.State)
	if err != nil {
		return v1.Service{}, err
	}

	message := status.Message

	var failureInfo *v1.ServiceFailureInfo
	if status.FailureInfo != nil {
		message = nil

		failureMessage := status.FailureInfo.Message
		if status.FailureInfo.Internal {
			failureMessage = "internal error"
		}

		failureInfo = &v1.ServiceFailureInfo{
			Time:    status.FailureInfo.Timestamp.Time,
			Message: failureMessage,
		}
	}

	// get service instances

	instances, err := kb.getServiceInstances(v1.ServiceID(id), namespace)

	if err != nil {
		return v1.Service{}, err
	}

	service := v1.Service{
		ID:   v1.ServiceID(id),
		Path: path,

		State:       state,
		Message:     message,
		FailureInfo: failureInfo,

		AvailableInstances:   status.AvailableInstances,
		UpdatedInstances:     status.UpdatedInstances,
		StaleInstances:       status.StaleInstances,
		TerminatingInstances: status.TerminatingInstances,

		Ports:     status.Ports,
		Instances: instances,
	}

	return service, nil
}

func getServiceState(state latticev1.ServiceState) (v1.ServiceState, error) {
	switch state {
	case latticev1.ServiceStatePending:
		return v1.ServiceStatePending, nil
	case latticev1.ServiceStateDeleting:
		return v1.ServiceStateDeleting, nil

	case latticev1.ServiceStateScaling:
		return v1.ServiceStateScaling, nil
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

func (kb *KubernetesBackend) getServiceInstances(serviceID v1.ServiceID, namespace string) ([]string, error) {

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{string(serviceID)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for service '%v' instances lookup: %v", serviceID, err)
	}

	selector = selector.Add(*requirement)
	pods, err := kb.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	instances := make([]string, len(pods.Items))

	for i, podItem := range pods.Items {
		instances[i] = toServiceInstanceShortID(serviceID, podItem.Name)
	}

	return instances, nil
}

func toServiceInstanceShortID(serviceID v1.ServiceID, podName string) string {
	// TODO Reuse existing deployment naming utilties
	return strings.TrimPrefix(podName, "lattice-service-"+string(serviceID)+"-")
}

func toServiceInstanceFullID(serviceID v1.ServiceID, podName string) string {
	// TODO Reuse existing deployment naming utilities
	return "lattice-service-" + string(serviceID) + "-" + podName
}
