package backend

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListComponentBuilds(systemID v1.SystemID) ([]v1.ComponentBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	buildList, err := kb.latticeClient.LatticeV1().ComponentBuilds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var builds []v1.ComponentBuild
	for _, build := range buildList.Items {
		builds = append(builds, transformComponentBuild(build.Name, build.Status))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetComponentBuild(
	systemID v1.SystemID,
	buildID v1.ComponentBuildID,
) (*v1.ComponentBuild, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(systemID, buildID)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalBuild := transformComponentBuild(build.Name, build.Status)
	return &externalBuild, true, nil
}

func (kb *KubernetesBackend) GetComponentBuildLogs(
	systemID v1.SystemID,
	buildID v1.ComponentBuildID,
	follow bool,
) (io.ReadCloser, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(systemID, buildID)
	if err != nil {
		return nil, false, err
	}

	if !exists {
		return nil, false, fmt.Errorf("ComponentBuild " + string(buildID) + " does not exist")
	}

	pod, err := kb.getPodForComponentBuild(build)
	if err != nil {
		return nil, false, err
	}

	if pod == nil {
		switch build.Status.State {
		case latticev1.ComponentBuildStatePending, latticev1.ComponentBuildStateQueued:
			return nil, false, fmt.Errorf("ComponentBuild " + string(buildID) + " not yet running")
		case latticev1.ComponentBuildStateRunning:
			return nil, false, fmt.Errorf("build for ComopnentBuild %v does not exist", buildID)
		case latticev1.ComponentBuildStateSucceeded, latticev1.ComponentBuildStateFailed:
			return nil, false, fmt.Errorf("ComponentBuild " + string(buildID) + " logs no longer available")
		default:
			panic("unreachable")
		}
	}

	req := kb.kubeClient.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &corev1.PodLogOptions{Follow: follow})

	readCloser, err := req.Stream()

	return readCloser, true, err
}

func (kb *KubernetesBackend) getInternalComponentBuild(
	systemID v1.SystemID,
	buildID v1.ComponentBuildID,
) (*latticev1.ComponentBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformComponentBuild(name string, status latticev1.ComponentBuildStatus) v1.ComponentBuild {
	var failureMessage *string
	if status.FailureInfo != nil {
		message := getComponentBuildFailureMessage(*status.FailureInfo)
		failureMessage = &message
	}

	externalBuild := v1.ComponentBuild{
		ID:                v1.ComponentBuildID(name),
		State:             getComponentBuildState(status.State),
		LastObservedPhase: status.LastObservedPhase,
		FailureMessage:    failureMessage,
	}

	return externalBuild
}

func getComponentBuildState(state latticev1.ComponentBuildState) v1.ComponentBuildState {
	switch state {
	case latticev1.ComponentBuildStatePending:
		return v1.ComponentBuildStatePending
	case latticev1.ComponentBuildStateQueued:
		return v1.ComponentBuildStateQueued
	case latticev1.ComponentBuildStateRunning:
		return v1.ComponentBuildStateRunning
	case latticev1.ComponentBuildStateSucceeded:
		return v1.ComponentBuildStateSucceeded
	case latticev1.ComponentBuildStateFailed:
		return v1.ComponentBuildStateFailed
	default:
		panic("unreachable")
	}
}

func getComponentBuildFailureMessage(failureInfo v1.ComponentBuildFailureInfo) string {
	if failureInfo.Internal {
		return "failed due to an internal error"
	}
	return failureInfo.Message
}

func (kb *KubernetesBackend) getPodForComponentBuild(build *latticev1.ComponentBuild) (*corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", kubeconstants.LabelKeyComponentBuildID, build.Name),
	}
	podsList, err := kb.kubeClient.CoreV1().Pods(build.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	pods := podsList.Items

	if len(pods) == 0 {
		return nil, nil
	}

	if len(pods) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Pods", build.Name)
	}

	return &pods[0], nil
}
