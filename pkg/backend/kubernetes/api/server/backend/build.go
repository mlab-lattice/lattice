package backend

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) Build(systemID v1.SystemID, definitionRoot tree.Node, version v1.SystemVersion) (*v1.Build, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	build, err := newBuild(definitionRoot, version)
	if err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	build, err = kb.latticeClient.LatticeV1().Builds(namespace).Create(build)
	if err != nil {
		return nil, err
	}

	externalBuild, err := transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func newBuild(definitionRoot tree.Node, version v1.SystemVersion) (*latticev1.Build, error) {
	labels := map[string]string{
		latticev1.BuildDefinitionVersionLabelKey: string(version),
	}

	services := map[tree.NodePath]latticev1.BuildSpecServiceInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = latticev1.BuildSpecServiceInfo{
			Definition: svcNode.Definition().(definition.Service),
		}
	}

	build := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.BuildSpec{
			DefinitionRoot: definitionRoot,
			Services:       services,
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
		externalBuild, err := transformBuild(&build)
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

	externalBuild, err := transformBuild(build)
	if err != nil {
		return nil, err
	}

	return &externalBuild, nil
}

func transformBuild(build *latticev1.Build) (v1.Build, error) {
	state, err := getBuildState(build.Status.State)
	if err != nil {
		return v1.Build{}, err
	}

	version := v1.SystemVersion("unknown")
	if label, ok := build.DefinitionVersionLabel(); ok {
		version = label
	}

	externalBuild := v1.Build{
		ID:       v1.BuildID(build.Name),
		State:    state,
		Version:  version,
		Services: map[tree.NodePath]v1.ServiceBuild{},
	}

	for service, serviceBuildName := range build.Status.ServiceBuilds {
		serviceBuildStatus, ok := build.Status.ServiceBuildStatuses[serviceBuildName]
		if !ok {
			err := fmt.Errorf(
				"System build %v/%v has ServiceBuild %v but no Status for it",
				build.Namespace,
				build.Name,
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

	externalBuild := v1.ServiceBuild{
		ID:         v1.ServiceBuildID(name),
		State:      state,
		Components: map[string]v1.ComponentBuild{},
	}

	for component, componentBuildName := range status.ComponentBuilds {
		componentBuildStatus, ok := status.ComponentBuildStatuses[componentBuildName]
		if !ok {
			err := fmt.Errorf(
				"ServiceBuild %v/%v has ComponentBuild %v for component %v but does not have its status",
				namespace,
				name,
				componentBuildName,
				component,
			)
			return v1.ServiceBuild{}, err
		}

		externalComponentBuild, err := transformComponentBuild(componentBuildName, componentBuildStatus)
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

func transformComponentBuild(name string, status latticev1.ComponentBuildStatus) (v1.ComponentBuild, error) {
	state, err := getComponentBuildState(status.State)
	if err != nil {
		return v1.ComponentBuild{}, err
	}

	var failureMessage *string
	if status.FailureInfo != nil {
		message := getComponentBuildFailureMessage(*status.FailureInfo)
		failureMessage = &message
	}

	externalBuild := v1.ComponentBuild{
		ID:                v1.ComponentBuildID(name),
		State:             state,
		LastObservedPhase: status.LastObservedPhase,
		FailureMessage:    failureMessage,
	}

	return externalBuild, nil
}

func getComponentBuildState(state latticev1.ComponentBuildState) (v1.ComponentBuildState, error) {
	switch state {
	case latticev1.ComponentBuildStatePending:
		return v1.ComponentBuildStatePending, nil
	case latticev1.ComponentBuildStateQueued:
		return v1.ComponentBuildStateQueued, nil
	case latticev1.ComponentBuildStateRunning:
		return v1.ComponentBuildStateRunning, nil
	case latticev1.ComponentBuildStateSucceeded:
		return v1.ComponentBuildStateSucceeded, nil
	case latticev1.ComponentBuildStateFailed:
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
