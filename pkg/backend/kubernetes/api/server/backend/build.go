package backend

import (
	"fmt"
	"io"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	defintionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) Build(
	systemID v1.SystemID,
	def *defintionv1.SystemNode,
	ri resolver.ResolutionInfo,
	version v1.SystemVersion,
) (*v1.Build, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	build, err := newBuild(def, ri, version)
	if err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	build, err = kb.latticeClient.LatticeV1().Builds(namespace).Create(build)
	if err != nil {
		return nil, err
	}

	externalBuild, err := kb.transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func newBuild(def *defintionv1.SystemNode, ri resolver.ResolutionInfo, version v1.SystemVersion) (*latticev1.Build, error) {
	labels := map[string]string{
		latticev1.BuildDefinitionVersionLabelKey: string(version),
	}

	build := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.BuildSpec{
			Definition:     def,
			ResolutionInfo: ri,
		},
	}

	return build, nil
}

func (kb *KubernetesBackend) ListBuilds(systemID v1.SystemID) ([]v1.Build, error) {
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	builds, err := kb.latticeClient.LatticeV1().Builds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// need to actually allocate the slice here so we return a slice instead of nil
	// if builds.Items is empty
	externalBuilds := make([]v1.Build, 0)
	for _, build := range builds.Items {
		externalBuild, err := kb.transformBuild(&build)
		if err != nil {
			return nil, err
		}

		externalBuilds = append(externalBuilds, externalBuild)
	}

	return externalBuilds, nil
}

func (kb *KubernetesBackend) GetBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Build, error) {
	// Ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	build, err := kb.latticeClient.LatticeV1().Builds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidBuildIDError(buildID)
		}

		return nil, err
	}

	externalBuild, err := kb.transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func (kb *KubernetesBackend) BuildLogs(
	systemID v1.SystemID,
	buildID v1.BuildID,
	path tree.Path,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	build, err := kb.latticeClient.LatticeV1().Builds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidBuildIDError(buildID)
		}

		return nil, err
	}

	service, ok := build.Status.Services[path]
	if !ok {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidServicePathError(path)
		}

		return nil, err
	}

	containerBuildID := service.MainContainer
	if sidecar != nil {
		containerBuildID, ok = service.Sidecars[*sidecar]
		if !ok {
			return nil, v1.NewInvalidSidecarError(*sidecar)
		}
	}

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildIDLabelKey, selection.Equals, []string{containerBuildID})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v/%v job lookup: %v", namespace, containerBuildID, err)
	}

	selector = selector.Add(*requirement)
	pods, err := kb.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("found multiple pods for %v/%v", namespace, containerBuildID)
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	pod := pods.Items[0]
	podLogOptions, err := toPodLogOptions(logOptions)
	if err != nil {
		return nil, err
	}
	req := kb.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, podLogOptions)
	return req.Stream()
}

func (kb *KubernetesBackend) transformBuild(build *latticev1.Build) (v1.Build, error) {
	state, err := getBuildState(build.Status.State)
	if err != nil {
		return v1.Build{}, err
	}

	version := v1.SystemVersion("unknown")
	if label, ok := build.DefinitionVersionLabel(); ok {
		version = label
	}

	var startTimestamp *time.Time
	if build.Status.StartTimestamp != nil {
		startTimestamp = &build.Status.StartTimestamp.Time
	}

	var completionTimestamp *time.Time
	if build.Status.CompletionTimestamp != nil {
		completionTimestamp = &build.Status.CompletionTimestamp.Time
	}

	externalBuild := v1.Build{
		ID:    v1.BuildID(build.Name),
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		Version:  version,
		Services: make(map[tree.Path]v1.ServiceBuild),
	}

	for path, serviceInfo := range build.Status.Services {
		externalServiceBuild, err := transformServiceBuild(
			build.Namespace,
			build.Name,
			&serviceInfo,
			build.Status.ContainerBuildStatuses,
		)
		if err != nil {
			return v1.Build{}, err
		}

		externalBuild.Services[path] = externalServiceBuild
	}

	return externalBuild, nil
}

func getBuildState(state latticev1.BuildState) (v1.BuildState, error) {
	switch state {
	case latticev1.BuildStatePending:
		return v1.BuildStatePending, nil
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

func transformServiceBuild(
	namespace, name string,
	serviceInfo *latticev1.BuildStatusService,
	containerBuildStatuses map[string]latticev1.ContainerBuildStatus,
) (v1.ServiceBuild, error) {
	mainContainerBuildStatus, ok := containerBuildStatuses[serviceInfo.MainContainer]
	if !ok {
		err := fmt.Errorf(
			"build %v/%v but does not have status for container build %v",
			namespace,
			name,
			serviceInfo.MainContainer,
		)
		return v1.ServiceBuild{}, err
	}

	externalContainerBuild, err := transformContainerBuild(serviceInfo.MainContainer, mainContainerBuildStatus)
	if err != nil {
		return v1.ServiceBuild{}, err
	}

	externalBuild := v1.ServiceBuild{
		ContainerBuild: externalContainerBuild,
		Sidecars:       make(map[string]v1.ContainerBuild),
	}

	for sidecar, containerBuildID := range serviceInfo.Sidecars {
		containerBuildStatus, ok := containerBuildStatuses[containerBuildID]
		if !ok {
			err := fmt.Errorf(
				"build %v/%v but does not have status for container build %v",
				namespace,
				name,
				serviceInfo.MainContainer,
			)
			return v1.ServiceBuild{}, err
		}

		externalContainerBuild, err := transformContainerBuild(containerBuildID, containerBuildStatus)
		if err != nil {
			return v1.ServiceBuild{}, err
		}

		externalBuild.Sidecars[sidecar] = externalContainerBuild
	}

	return externalBuild, nil
}

func transformContainerBuild(containerBuildID string, status latticev1.ContainerBuildStatus) (v1.ContainerBuild, error) {
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
		startTimestamp = &status.StartTimestamp.Time
	}

	var completionTimestamp *time.Time
	if status.CompletionTimestamp != nil {
		completionTimestamp = &status.CompletionTimestamp.Time
	}

	externalBuild := v1.ContainerBuild{
		ID:    v1.ContainerBuildID(containerBuildID),
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		LastObservedPhase: phase,
		FailureMessage:    failureMessage,
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
