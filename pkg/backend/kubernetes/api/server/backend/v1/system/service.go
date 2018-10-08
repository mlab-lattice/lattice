package system

import (
	"fmt"
	"io"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/time"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type serviceBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *serviceBackend) List() ([]v1.Service, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	services, err := b.backend.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []v1.Service
	for _, service := range services.Items {
		servicePath, err := service.PathLabel()
		if err != nil {
			return nil, err
		}

		externalService, err := b.transformService(v1.ServiceID(service.Name), servicePath, &service, namespace)
		if err != nil {
			return nil, err
		}

		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (b *serviceBackend) Get(id v1.ServiceID) (*v1.Service, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	service, err := b.backend.latticeClient.LatticeV1().Services(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	servicePath, err := service.PathLabel()

	if err != nil {
		return nil, err
	}

	externalService, err := b.transformService(v1.ServiceID(service.Name), servicePath, service, namespace)
	if err != nil {
		return nil, err
	}

	return &externalService, nil
}

func (b *serviceBackend) GetByPath(path tree.Path) (*v1.Service, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})

	if err != nil {
		return nil, fmt.Errorf("error getting selector for  service %v in namespace %v", path.String(), namespace)
	}

	selector = selector.Add(*requirement)

	services, err := b.backend.latticeClient.LatticeV1().Services(namespace).List(
		metav1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		return nil, err
	}

	if len(services.Items) == 0 {
		return nil, nil
	}

	// ensure result is a singleton

	if len(services.Items) > 1 {
		return nil, fmt.Errorf("found multiple services with path %v in namespace %v", path.String(), namespace)
	}

	service := services.Items[0]
	externalService, err := b.transformService(v1.ServiceID(service.Name), path, &service, namespace)
	if err != nil {
		return nil, err
	}

	return &externalService, nil

}
func (b *serviceBackend) Logs(
	id v1.ServiceID,
	sidecar *string,
	instance string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)

	_, err := b.backend.latticeClient.LatticeV1().Services(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pod, err := b.findServicePod(id, instance, namespace)

	if err != nil {
		return nil, err
	}

	container := kubeutil.UserMainContainerName
	if sidecar != nil {
		container = kubeutil.UserSidecarContainerName(*sidecar)
	}

	podLogOptions, err := toPodLogOptions(logOptions)
	if err != nil {
		return nil, err
	}
	podLogOptions.Container = container

	req := b.backend.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, podLogOptions)
	return req.Stream()

}

// findServicePod finds service pod by instance id or service's single pod if id was not specified
func (b *serviceBackend) findServicePod(serviceId v1.ServiceID, instance string, namespace string) (*corev1.Pod, error) {

	// check if instance was specified
	if instance != "" {
		podName := toServiceInstanceFullID(serviceId, instance)
		pod, err := b.backend.kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
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
	pods, err := b.backend.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
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

func (b *serviceBackend) transformService(
	id v1.ServiceID,
	path tree.Path,
	service *latticev1.Service,
	namespace string,
) (v1.Service, error) {
	state, err := getServiceState(service.Status.State)
	if err != nil {
		return v1.Service{}, err
	}

	message := service.Status.Message

	var failureInfo *v1.ServiceFailureInfo
	if service.Status.FailureInfo != nil {
		message = nil

		failureMessage := service.Status.FailureInfo.Message
		if service.Status.FailureInfo.Internal {
			failureMessage = "internal error"
		}

		failureInfo = &v1.ServiceFailureInfo{
			Time:    *time.New(service.Status.FailureInfo.Timestamp.Time),
			Message: failureMessage,
		}
	}

	// get service instances
	instances, err := b.getServiceInstances(id, namespace)

	if err != nil {
		return v1.Service{}, err
	}

	// get service instance metrics
	instanceMetrics, err := b.getServiceInstanceMetrics(id, namespace)

	if err != nil {
		return v1.Service{}, err
	}

	externalService := v1.Service{
		ID: v1.ServiceID(id),

		Path: path,

		Status: v1.ServiceStatus{
			State:       state,
			Message:     message,
			FailureInfo: failureInfo,

			AvailableInstances:   service.Status.AvailableInstances,
			UpdatedInstances:     service.Status.UpdatedInstances,
			StaleInstances:       service.Status.StaleInstances,
			TerminatingInstances: service.Status.TerminatingInstances,

			Ports:           service.Status.Ports,
			Instances:       instances,
			InstanceMetrics: instanceMetrics,
		},
	}
	return externalService, nil
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

func (b *serviceBackend) getServiceInstances(id v1.ServiceID, namespace string) ([]string, error) {

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{string(id)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for service '%v' instances lookup: %v", id, err)
	}

	selector = selector.Add(*requirement)
	pods, err := b.backend.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	instances := make([]string, len(pods.Items))

	for i, podItem := range pods.Items {
		instances[i] = toServiceInstanceShortID(id, podItem.Name)
	}

	return instances, nil
}

func (b *serviceBackend) getServiceInstanceMetrics(id v1.ServiceID, namespace string) ([]string, error) {

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{string(id)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for service '%v' instances lookup: %v", id, err)
	}

	selector = selector.Add(*requirement)
	pods, err := b.backend.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	instances := make([]string, len(pods.Items))

	for i, podItem := range pods.Items {
		instances[i] = podItem.Name
		//instances[i] = toServiceInstanceShortID(id, podItem.Name)
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
