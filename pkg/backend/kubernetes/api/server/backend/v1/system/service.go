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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type serviceBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *serviceBackend) namespace() string {
	return b.backend.systemNamespace(b.system)
}

func (b *serviceBackend) List() ([]v1.Service, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	services, err := b.backend.latticeClient.LatticeV1().Services(b.namespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	externalServices := make([]v1.Service, len(services.Items))
	for i := 0; i < len(services.Items); i++ {
		externalService, err := b.transformService(&services.Items[i])
		if err != nil {
			return nil, err
		}

		externalServices[i] = externalService
	}

	return externalServices, nil
}

func (b *serviceBackend) Get(id v1.ServiceID) (*v1.Service, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	service, err := b.backend.latticeClient.LatticeV1().Services(b.namespace()).Get(string(id), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	externalService, err := b.transformService(service)
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

	service, ok, err := b.serviceForPath(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	externalService, err := b.transformService(service)
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

	pod, ok, err := b.pod(id, instance)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, v1.NewInvalidServiceInstanceIDError()
	}

	container := kubeutil.UserMainContainerName
	if sidecar != nil {
		container = kubeutil.UserSidecarContainerName(*sidecar)
	}

	podLogOptions, err := toPodLogOptions(logOptions, container)
	if err != nil {
		return nil, err
	}

	// FIXME(kevindrosendahl): factor out and include podLogsShouldBeAvailable
	req := b.backend.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, podLogOptions)
	return req.Stream()

}

func (b *serviceBackend) serviceForPath(path tree.Path) (*latticev1.Service, bool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, false, fmt.Errorf("error getting selector for  service %v in namespace %v", path.String(), b.namespace())
	}

	selector = selector.Add(*requirement)

	options := metav1.ListOptions{LabelSelector: selector.String()}
	services, err := b.backend.latticeClient.LatticeV1().Services(b.namespace()).List(options)
	if err != nil {
		return nil, false, err
	}

	if len(services.Items) == 0 {
		return nil, false, nil
	}

	// ensure result is a singleton

	if len(services.Items) > 1 {
		return nil, false, fmt.Errorf("found multiple services with path %v in namespace %v", path.String(), b.namespace())
	}

	return &services.Items[0], true, nil
}

func (b *serviceBackend) transformService(service *latticev1.Service) (v1.Service, error) {
	path, err := service.PathLabel()
	if err != nil {
		return v1.Service{}, err
	}

	state, err := getServiceState(service.Status.State)
	if err != nil {
		return v1.Service{}, err
	}

	id := v1.ServiceID(service.Name)
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

	// get service pods
	pods, err := b.pods(service)
	if err != nil {
		return v1.Service{}, err
	}

	instances := make([]string, len(pods.Items))
	for i := 0; i < len(pods.Items); i++ {
		instances[i] = toServiceInstanceShortID(id, pods.Items[i].Name)
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

			Ports:     service.Status.Ports,
			Instances: instances,
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

func (b *serviceBackend) pods(service *latticev1.Service) (*corev1.PodList, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{string(service.Name)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for service '%v' pods lookup: %v", service.Name, err)
	}

	selector = selector.Add(*requirement)
	return b.backend.kubeClient.CoreV1().Pods(service.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
}

// pod finds service pod by instance id or service's single pod if id was not specified
func (b *serviceBackend) pod(id v1.ServiceID, instance string) (*corev1.Pod, bool, error) {
	podName := toServiceInstanceFullID(id, instance)
	pod, err := b.backend.kubeClient.CoreV1().Pods(b.namespace()).Get(podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return pod, true, nil
}

func toServiceInstanceShortID(serviceID v1.ServiceID, podName string) string {
	// TODO Reuse existing deployment naming utilties
	return strings.TrimPrefix(podName, "lattice-service-"+string(serviceID)+"-")
}

func toServiceInstanceFullID(serviceID v1.ServiceID, podName string) string {
	// TODO Reuse existing deployment naming utilities
	return "lattice-service-" + string(serviceID) + "-" + podName
}
