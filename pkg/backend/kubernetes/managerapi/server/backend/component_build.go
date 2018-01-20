package backend

import (
	"fmt"
	"io"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	backend "github.com/mlab-lattice/system/pkg/managerapi/server"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListComponentBuilds(id types.SystemID) ([]types.ComponentBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	buildList, err := kb.LatticeClient.LatticeV1().ComponentBuilds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var builds []types.ComponentBuild
	for _, build := range buildList.Items {
		builds = append(builds, transformComponentBuild(build.Name, build.Status))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetComponentBuild(id types.SystemID, bid types.ComponentBuildID) (*types.ComponentBuild, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(id, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalBuild := transformComponentBuild(build.Name, build.Status)
	return &externalBuild, true, nil
}

func (kb *KubernetesBackend) GetComponentBuildLogs(id types.SystemID, bid types.ComponentBuildID, follow bool) (io.ReadCloser, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(id, bid)
	if err != nil {
		return nil, false, err
	}

	if !exists {
		return nil, false, backend.NewUserError("ComponentBuild " + string(bid) + " does not exist")
	}

	pod, err := kb.getPodForComponentBuild(build)
	if err != nil {
		return nil, false, err
	}

	if pod == nil {
		switch build.Status.State {
		case crv1.ComponentBuildStatePending, crv1.ComponentBuildStateQueued:
			return nil, false, backend.NewUserError("ComponentBuild " + string(bid) + " not yet running")
		case crv1.ComponentBuildStateRunning:
			return nil, false, fmt.Errorf("build for ComopnentBuild %v does not exist", bid)
		case crv1.ComponentBuildStateSucceeded, crv1.ComponentBuildStateFailed:
			return nil, false, backend.NewUserError("ComponentBuild " + string(bid) + " logs no longer available")
		default:
			panic("unreachable")
		}
	}

	req := kb.KubeClient.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &corev1.PodLogOptions{Follow: follow})

	readCloser, err := req.Stream()

	return readCloser, true, err
}

func (kb *KubernetesBackend) getInternalComponentBuild(id types.SystemID, bid types.ComponentBuildID) (*crv1.ComponentBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformComponentBuild(name string, status crv1.ComponentBuildStatus) types.ComponentBuild {
	var failureMessage *string
	if status.FailureInfo != nil {
		message := getComponentBuildFailureMessage(*status.FailureInfo)
		failureMessage = &message
	}

	externalBuild := types.ComponentBuild{
		ID:                types.ComponentBuildID(name),
		State:             getComponentBuildState(status.State),
		LastObservedPhase: status.LastObservedPhase,
		FailureMessage:    failureMessage,
	}

	return externalBuild
}

func getComponentBuildState(state crv1.ComponentBuildState) types.ComponentBuildState {
	switch state {
	case crv1.ComponentBuildStatePending:
		return types.ComponentBuildStatePending
	case crv1.ComponentBuildStateQueued:
		return types.ComponentBuildStateQueued
	case crv1.ComponentBuildStateRunning:
		return types.ComponentBuildStateRunning
	case crv1.ComponentBuildStateSucceeded:
		return types.ComponentBuildStateSucceeded
	case crv1.ComponentBuildStateFailed:
		return types.ComponentBuildStateFailed
	default:
		panic("unreachable")
	}
}

func getComponentBuildFailureMessage(failureInfo types.ComponentBuildFailureInfo) string {
	if failureInfo.Internal {
		return "failed due to an internal error"
	}
	return failureInfo.Message
}

func (kb *KubernetesBackend) getPodForComponentBuild(cb *crv1.ComponentBuild) (*corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", kubeconstants.LabelKeyComponentBuildID, cb.Name),
	}
	podsList, err := kb.KubeClient.CoreV1().Pods(cb.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	pods := podsList.Items

	if len(pods) == 0 {
		return nil, nil
	}

	if len(pods) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Pods", cb.Name)
	}

	return &pods[0], nil
}
