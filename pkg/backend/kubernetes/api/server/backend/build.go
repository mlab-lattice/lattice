package backend

import (
	"fmt"
	"io"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) Build(systemID v1.SystemID, def *tree.SystemNode, version v1.SystemVersion) (*v1.Build, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	build, err := newBuild(def, version)
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

func newBuild(def *tree.SystemNode, version v1.SystemVersion) (*latticev1.Build, error) {
	labels := map[string]string{
		latticev1.BuildDefinitionVersionLabelKey: string(version),
	}

	services := make(map[tree.NodePath]latticev1.BuildSpecServiceInfo)
	for path, svcNode := range def.Services() {
		services[path] = latticev1.BuildSpecServiceInfo{
			Definition: svcNode.Definition().(*definition.Service),
		}
	}

	build := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.BuildSpec{
			Definition: def,
			Services:   services,
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
	path tree.NodePath,
	component string,
	follow bool,
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

	serviceBuildID, ok := build.Status.ContainerBuilds[path]
	if !ok {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidServicePathError(path)
		}

		return nil, err
	}

	status, ok := build.Status.ContainerBuildStatuses[serviceBuildID]
	if !ok {
		err := fmt.Errorf(
			"%v has service build ID %v for %v, but does not have a status for it",
			build.Description(kb.namespacePrefix),
			serviceBuildID,
			path.String(),
		)
		return nil, err
	}

	componentBuildID, ok := status.ComponentBuilds[component]
	if !ok {
		return nil, v1.NewInvalidComponentError(component)
	}

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildIDLabelKey, selection.Equals, []string{componentBuildID})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v/%v job lookup: %v", namespace, componentBuildID, err)
	}

	selector = selector.Add(*requirement)
	pods, err := kb.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("found multiple pods for %v/%v", namespace, componentBuildID)
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	pod := pods.Items[0]
	logOptions := &corev1.PodLogOptions{Follow: follow}
	req := kb.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, logOptions)
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
		Services: map[tree.NodePath]v1.ServiceBuild{},
	}

	for service, serviceBuildName := range build.Status.ContainerBuilds {
		serviceBuildStatus, ok := build.Status.ContainerBuildStatuses[serviceBuildName]
		if !ok {
			err := fmt.Errorf(
				"%v has service build %v but no Status for it",
				build.Description(kb.namespacePrefix),
				serviceBuildName,
			)
			return v1.Build{}, err
		}

		externalServiceBuild, err := transformServiceBuild(build.Namespace, serviceBuildName, &serviceBuildStatus)
		if err != nil {
			return v1.Build{}, err
		}

		externalBuild.Services[service] = externalServiceBuild
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

func transformServiceBuild(namespace, name string, status *latticev1.ServiceBuildStatus) (v1.ServiceBuild, error) {
	state, err := getServiceBuildState(status.State)
	if err != nil {
		return v1.ServiceBuild{}, err
	}

	var startTimestamp *time.Time
	if status.StartTimestamp != nil {
		startTimestamp = &status.StartTimestamp.Time
	}

	var completionTimestamp *time.Time
	if status.CompletionTimestamp != nil {
		completionTimestamp = &status.CompletionTimestamp.Time
	}

	externalBuild := v1.ServiceBuild{
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		Components: map[string]v1.ComponentBuild{},
	}

	for component, componentBuildName := range status.ComponentBuilds {
		componentBuildStatus, ok := status.ComponentBuildStatuses[componentBuildName]
		if !ok {
			err := fmt.Errorf(
				"service build %v/%v has component build %v for component %v but does not have its status",
				namespace,
				name,
				componentBuildName,
				component,
			)
			return v1.ServiceBuild{}, err
		}

		externalComponentBuild, err := transformComponentBuild(componentBuildStatus)
		if err != nil {
			return v1.ServiceBuild{}, err
		}

		externalBuild.Components[component] = externalComponentBuild
	}

	return externalBuild, nil
}

func getServiceBuildState(state latticev1.ServiceBuildState) (v1.ServiceBuildState, error) {
	switch state {
	case latticev1.ServiceBuildStatePending:
		return v1.ServiceBuildStatePending, nil
	case latticev1.ServiceBuildStateRunning:
		return v1.ServiceBuildStateRunning, nil
	case latticev1.ServiceBuildStateSucceeded:
		return v1.ServiceBuildStateSucceeded, nil
	case latticev1.ServiceBuildStateFailed:
		return v1.ServiceBuildStateFailed, nil
	default:
		return "", fmt.Errorf("invalid service build state: %v", state)
	}
}

func transformComponentBuild(status latticev1.ContainerBuildStatus) (v1.ComponentBuild, error) {
	state, err := getComponentBuildState(status.State)
	if err != nil {
		return v1.ComponentBuild{}, err
	}

	var failureMessage *string
	if status.FailureInfo != nil {
		message := getComponentBuildFailureMessage(*status.FailureInfo)
		failureMessage = &message
	}

	phase := status.LastObservedPhase
	if state == v1.ComponentBuildStateSucceeded {
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

	externalBuild := v1.ComponentBuild{
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		LastObservedPhase: phase,
		FailureMessage:    failureMessage,
	}

	return externalBuild, nil
}

func getComponentBuildState(state latticev1.ComponentBuildState) (v1.ComponentBuildState, error) {
	switch state {
	case latticev1.ContainerBuildStatePending:
		return v1.ComponentBuildStatePending, nil
	case latticev1.ContainerBuildStateQueued:
		return v1.ComponentBuildStateQueued, nil
	case latticev1.ContainerBuildStateRunning:
		return v1.ComponentBuildStateRunning, nil
	case latticev1.ContainerBuildStateSucceeded:
		return v1.ComponentBuildStateSucceeded, nil
	case latticev1.ContainerBuildStateFailed:
		return v1.ComponentBuildStateFailed, nil
	default:
		return "", fmt.Errorf("invalid component state: %v", state)
	}
}

func getComponentBuildFailureMessage(failureInfo v1.ComponentBuildFailureInfo) string {
	if failureInfo.Internal {
		return "failed due to an internal error"
	}
	return failureInfo.Message
}
