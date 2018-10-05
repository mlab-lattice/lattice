package system

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/time"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

type buildBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *buildBackend) namespace() string {
	return b.backend.systemNamespace(b.system)
}

func (b *buildBackend) CreateFromVersion(version v1.Version) (*v1.Build, error) {
	return b.createBuild(&version, nil)
}

func (b *buildBackend) CreateFromPath(path tree.Path) (*v1.Build, error) {
	return b.createBuild(nil, &path)
}

func (b *buildBackend) createBuild(version *v1.Version, path *tree.Path) (*v1.Build, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	build := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: map[string]string{},
		},
		Spec: latticev1.BuildSpec{
			Version: version,
			Path:    path,
		},
	}

	build, err := b.backend.latticeClient.LatticeV1().Builds(b.namespace()).Create(build)
	if err != nil {
		return nil, err
	}

	externalBuild, err := b.transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func (b *buildBackend) List() ([]v1.Build, error) {
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	builds, err := b.backend.latticeClient.LatticeV1().Builds(b.namespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	externalBuilds := make([]v1.Build, len(builds.Items))
	for i := 0; i < len(builds.Items); i++ {
		externalBuild, err := b.transformBuild(&builds.Items[i])
		if err != nil {
			return nil, err
		}

		externalBuilds[i] = externalBuild
	}

	return externalBuilds, nil
}

func (b *buildBackend) Get(id v1.BuildID) (*v1.Build, error) {
	// Ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	build, err := b.backend.latticeClient.LatticeV1().Builds(b.namespace()).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidBuildIDError()
		}

		return nil, err
	}

	externalBuild, err := b.transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func (b *buildBackend) Logs(
	id v1.BuildID,
	path tree.Path,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	build, err := b.backend.latticeClient.LatticeV1().Builds(b.namespace()).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidBuildIDError()
		}

		return nil, err
	}

	workload, ok := build.Status.Workloads[path]
	if !ok {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidPathError()
		}

		return nil, err
	}

	containerBuildID := workload.MainContainer
	if sidecar != nil {
		containerBuildID, ok = workload.Sidecars[*sidecar]
		if !ok {
			return nil, v1.NewInvalidSidecarError()
		}
	}

	pod, ok, err := b.buildPod(containerBuildID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, v1.NewLogsUnavailableError()
	}

	podLogOptions, err := toPodLogOptions(logOptions, "")
	if err != nil {
		return nil, err
	}

	req := b.backend.kubeClient.CoreV1().Pods(b.namespace()).GetLogs(pod.Name, podLogOptions)
	return req.Stream()
}

func (b *buildBackend) buildPod(id v1.ContainerBuildID) (*corev1.Pod, bool, error) {
	namespace := b.namespace()

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildIDLabelKey, selection.Equals, []string{string(id)})
	if err != nil {
		err := fmt.Errorf("error creating requirement for container build %v/%v job lookup: %v", namespace, id, err)
		return nil, false, err
	}

	selector = selector.Add(*requirement)
	pods, err := b.backend.kubeClient.CoreV1().Pods(b.namespace()).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, false, err
	}

	if len(pods.Items) > 1 {
		err := fmt.Errorf("found multiple pods for container build %v/%v job lookup: %v", namespace, id, err)
		return nil, false, err
	}

	if len(pods.Items) == 0 {
		return nil, false, nil
	}

	return &pods.Items[0], true, nil
}

func (b *buildBackend) transformBuild(build *latticev1.Build) (v1.Build, error) {
	state, err := getBuildState(build.Status.State)
	if err != nil {
		return v1.Build{}, err
	}

	var startTimestamp *time.Time
	if build.Status.StartTimestamp != nil {
		startTimestamp = time.New(build.Status.StartTimestamp.Time)
	}

	var completionTimestamp *time.Time
	if build.Status.CompletionTimestamp != nil {
		completionTimestamp = time.New(build.Status.CompletionTimestamp.Time)
	}

	externalBuild := v1.Build{
		ID: v1.BuildID(build.Name),

		Path:    build.Spec.Path,
		Version: build.Spec.Version,

		Status: v1.BuildStatus{
			State:   state,
			Message: build.Status.Message,

			StartTimestamp:      startTimestamp,
			CompletionTimestamp: completionTimestamp,

			Path:    build.Status.Path,
			Version: build.Status.Version,

			Workloads: make(map[tree.Path]v1.WorkloadBuild),
		},
	}

	for path, workload := range build.Status.Workloads {
		externalServiceBuild, err := transformWorkloadBuild(
			build.Namespace,
			build.Name,
			&workload,
			build.Status.ContainerBuildStatuses,
		)
		if err != nil {
			return v1.Build{}, err
		}

		externalBuild.Status.Workloads[path] = externalServiceBuild
	}

	return externalBuild, nil
}

func getBuildState(state latticev1.BuildState) (v1.BuildState, error) {
	switch state {
	case latticev1.BuildStatePending:
		return v1.BuildStatePending, nil

	case latticev1.BuildStateAccepted:
		return v1.BuildStateAccepted, nil

	case latticev1.BuildStateRunning:
		return v1.BuildStateRunning, nil

	case latticev1.BuildStateSucceeded:
		return v1.BuildStateSucceeded, nil

	case latticev1.BuildStateFailed:
		return v1.BuildStateFailed, nil

	default:
		return "", fmt.Errorf("invalid build state: %v", state)
	}
}

func transformWorkloadBuild(
	namespace, name string,
	workload *latticev1.BuildStatusWorkload,
	containerBuildStatuses map[v1.ContainerBuildID]latticev1.ContainerBuildStatus,
) (v1.WorkloadBuild, error) {
	mainContainerBuildStatus, ok := containerBuildStatuses[workload.MainContainer]
	if !ok {
		err := fmt.Errorf(
			"build %v/%v but does not have status for container build %v",
			namespace,
			name,
			workload.MainContainer,
		)
		return v1.WorkloadBuild{}, err
	}

	externalContainerBuild, err := transformContainerBuild(workload.MainContainer, mainContainerBuildStatus)
	if err != nil {
		return v1.WorkloadBuild{}, err
	}

	externalBuild := v1.WorkloadBuild{
		ContainerBuild: externalContainerBuild,
		Sidecars:       make(map[string]v1.ContainerBuild),
	}

	for sidecar, containerBuildID := range workload.Sidecars {
		containerBuildStatus, ok := containerBuildStatuses[containerBuildID]
		if !ok {
			err := fmt.Errorf(
				"build %v/%v but does not have status for container build %v",
				namespace,
				name,
				workload.MainContainer,
			)
			return v1.WorkloadBuild{}, err
		}

		externalContainerBuild, err := transformContainerBuild(containerBuildID, containerBuildStatus)
		if err != nil {
			return v1.WorkloadBuild{}, err
		}

		externalBuild.Sidecars[sidecar] = externalContainerBuild
	}

	return externalBuild, nil
}

func transformContainerBuild(id v1.ContainerBuildID, status latticev1.ContainerBuildStatus) (v1.ContainerBuild, error) {
	state, err := getComponentBuildState(status.State)
	if err != nil {
		return v1.ContainerBuild{}, err
	}

	var failureMessage *string
	if status.FailureInfo != nil {
		message := getComponentBuildFailureMessage(*status.FailureInfo)
		failureMessage = &message
	}

	phase := status.LastObservedPhase
	if state == v1.ContainerBuildStateSucceeded {
		phase = nil
	}

	var startTimestamp *time.Time
	if status.StartTimestamp != nil {
		startTimestamp = time.New(status.StartTimestamp.Time)
	}

	var completionTimestamp *time.Time
	if status.CompletionTimestamp != nil {
		completionTimestamp = time.New(status.CompletionTimestamp.Time)
	}

	externalBuild := v1.ContainerBuild{
		ID: id,

		Status: v1.ContainerBuildStatus{
			State: state,

			StartTimestamp:      startTimestamp,
			CompletionTimestamp: completionTimestamp,

			LastObservedPhase: phase,
			FailureMessage:    failureMessage,
		},
	}

	return externalBuild, nil
}

func getComponentBuildState(state latticev1.ComponentBuildState) (v1.ContainerBuildState, error) {
	switch state {
	case latticev1.ContainerBuildStatePending:
		return v1.ContainerBuildStatePending, nil
	case latticev1.ContainerBuildStateQueued:
		return v1.ContainerBuildStateQueued, nil
	case latticev1.ContainerBuildStateRunning:
		return v1.ContainerBuildStateRunning, nil
	case latticev1.ContainerBuildStateSucceeded:
		return v1.ContainerBuildStateSucceeded, nil
	case latticev1.ContainerBuildStateFailed:
		return v1.ContainerBuildStateFailed, nil
	default:
		return "", fmt.Errorf("invalid component state: %v", state)
	}
}

func getComponentBuildFailureMessage(failureInfo v1.ContainerBuildFailureInfo) string {
	if failureInfo.Internal {
		return "failed due to an internal error"
	}
	return failureInfo.Message
}
