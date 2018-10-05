package system

import (
	"fmt"
	"io"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/time"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type jobRunBackend struct {
	backend *Backend
	system  v1.SystemID
	job     v1.JobID
}

func (b *jobRunBackend) namespace() string {
	return b.backend.systemNamespace(b.system)
}

func (b *jobRunBackend) List() ([]v1.JobRun, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	pods, err := b.getJobPods()
	if err != nil {
		return nil, err
	}

	var externalJobsRuns []v1.JobRun
	for _, pod := range pods {
		externalJobRun, err := b.transformJobRun(&pod)
		if err != nil {
			return nil, err
		}

		externalJobsRuns = append(externalJobsRuns, externalJobRun)
	}

	return externalJobsRuns, nil
}

func (b *jobRunBackend) Get(id v1.JobRunID) (*v1.JobRun, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	pod, err := b.getJobRunPod(id)
	if err != nil {
		return nil, err
	}

	externalJob, err := b.transformJobRun(pod)
	if err != nil {
		return nil, err
	}

	return &externalJob, nil
}

func (b *jobRunBackend) Logs(
	id v1.JobRunID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	pod, err := b.getJobRunPod(id)
	if err != nil {
		return nil, err
	}

	container := kubeutil.UserMainContainerName
	if sidecar != nil {
		container = kubeutil.UserSidecarContainerName(*sidecar)
	}

	logsAvailable, err := podLogsShouldBeAvailable(pod, container)
	if err != nil {
		return nil, err
	}

	if !logsAvailable {
		return nil, v1.NewLogsUnavailableError()
	}

	podLogOptions, err := toPodLogOptions(logOptions, container)
	if err != nil {
		return nil, err
	}

	req := b.backend.kubeClient.CoreV1().Pods(b.namespace()).GetLogs(pod.Name, podLogOptions)
	return req.Stream()
}

// pod finds service pod by instance id or service's single pod if id was not specified
func (b *jobRunBackend) getJobPods() ([]corev1.Pod, error) {
	namespace := b.backend.systemNamespace(b.system)

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobIDLabelKey, selection.Equals, []string{string(b.job)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v/%v job lookup: %v", namespace, b.job, err)
	}

	selector = selector.Add(*requirement)
	pods, err := b.backend.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (b *jobRunBackend) getJobRunPod(id v1.JobRunID) (*corev1.Pod, error) {
	podName := fmt.Sprintf("lattice-job-%v-%v", b.job, id)
	namespace := b.backend.systemNamespace(b.system)

	pod, err := b.backend.kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidJobRunIDError()
		}

		return nil, err
	}

	return pod, nil
}

func (b *jobRunBackend) transformJobRun(pod *corev1.Pod) (v1.JobRun, error) {
	status, err := jobRunStatus(pod)
	if err != nil {
		return v1.JobRun{}, err
	}

	run := v1.JobRun{
		ID: b.jobRunID(pod),

		Status: status,
	}
	return run, nil
}

func (b *jobRunBackend) jobRunID(pod *corev1.Pod) v1.JobRunID {
	return v1.JobRunID(strings.TrimPrefix(pod.Name, fmt.Sprintf("lattice-job-%v-", string(b.job))))

}

func jobRunStatus(pod *corev1.Pod) (v1.JobRunStatus, error) {
	var startTimestamp *time.Time
	if pod.Status.StartTime != nil {
		startTimestamp = time.New(pod.Status.StartTime.Time)
	}

	var completionTimestamp *time.Time
	var state v1.JobRunState
	switch pod.Status.Phase {
	case corev1.PodPending:
		state = v1.JobRunStatePending

	case corev1.PodRunning:
		state = v1.JobRunStateRunning

	case corev1.PodSucceeded:
		state = v1.JobRunStateSucceeded

	case corev1.PodFailed:
		state = v1.JobRunStateFailed

		var err error
		completionTimestamp, err = jobRunCompletionTimestamp(pod)
		if err != nil {
			return v1.JobRunStatus{}, err
		}

	default:
		state = v1.JobRunStateUnknown

		var err error
		completionTimestamp, err = jobRunCompletionTimestamp(pod)
		if err != nil {
			return v1.JobRunStatus{}, err
		}
	}

	status := v1.JobRunStatus{
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,
	}
	return status, nil
}

func jobRunCompletionTimestamp(pod *corev1.Pod) (*time.Time, error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name != kubeutil.UserMainContainerName {
			continue
		}

		if status.State.Terminated == nil {
			return nil, fmt.Errorf("container %v does not have terminated state", kubeutil.UserMainContainerName)
		}

		return time.New(status.State.Terminated.FinishedAt.Time), nil
	}

	return nil, fmt.Errorf("pod does not have status for %v", kubeutil.UserMainContainerName)
}

// the logs endpoint will return a BadRequest error when trying tail the logs of a
// container that is still creating. so before making the request check to see if
// we should get a good response. this way we can simple return a BadRequestError
// if one is returned (meaning there's a bug in this code) rather than trying to parse
// whether the bac request was due to the container still creating
func podLogsShouldBeAvailable(pod *corev1.Pod, container string) (bool, error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name != container {
			continue
		}

		running := status.State.Running != nil
		terminated := status.State.Terminated != nil
		return running || terminated, nil
	}

	if container == kubeutil.UserMainContainerName {
		return false, fmt.Errorf("pod does not have status for %v", kubeutil.UserMainContainerName)
	}

	return false, v1.NewInvalidSidecarError()
}
