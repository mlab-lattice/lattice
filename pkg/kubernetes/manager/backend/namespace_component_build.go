package backend

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/constants"
	kubeconstants "github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	"github.com/mlab-lattice/system/pkg/manager/backend"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListComponentBuilds(ln types.LatticeNamespace) ([]types.ComponentBuild, error) {
	result := &crv1.ComponentBuildList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(kubeconstants.NamespaceLatticeInternal).
		Resource(crv1.ResourcePluralComponentBuild).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	builds := []types.ComponentBuild{}
	for _, build := range result.Items {
		// FIXME: should add a label to component builds for the lattice namespace
		builds = append(builds, transformComponentBuild(&build))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetComponentBuild(ln types.LatticeNamespace, bid types.ComponentBuildID) (*types.ComponentBuild, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	coreBuild := transformComponentBuild(build)
	// FIXME: should add a label to component builds for the lattice namespace
	return &coreBuild, true, nil
}

func (kb *KubernetesBackend) GetComponentBuildLogs(ln types.LatticeNamespace, bid types.ComponentBuildID, follow bool) (io.ReadCloser, bool, error) {
	build, exists, err := kb.getInternalComponentBuild(ln, bid)
	if err != nil {
		return nil, false, err
	}

	if !exists {
		return nil, false, backend.NewUserError("ComponentBuild " + string(bid) + " does not exist")
	}

	pod, err := kb.getPodForComponentBuild(build)
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

	req := kb.KubeClientset.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &corev1.PodLogOptions{Follow: follow})

	readCloser, err := req.Stream()

	return readCloser, true, err
}

func (kb *KubernetesBackend) getInternalComponentBuild(ln types.LatticeNamespace, bid types.ComponentBuildID) (*crv1.ComponentBuild, bool, error) {
	result := &crv1.ComponentBuild{}
	err := kb.LatticeResourceClient.Get().
		Namespace(kubeconstants.NamespaceLatticeInternal).
		Resource(crv1.ResourcePluralComponentBuild).
		Name(string(bid)).
		Do().
		Into(result)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformComponentBuild(build *crv1.ComponentBuild) types.ComponentBuild {
	cb := types.ComponentBuild{
		ID:                types.ComponentBuildID(build.Name),
		State:             getComponentBuildState(build.Status.State),
		LastObservedPhase: build.Status.LastObservedPhase,
	}

	if build.Status.FailureInfo != nil {
		failureMessage := getComponentBuildFailureMessage(*build.Status.FailureInfo)
		cb.FailureMessage = &failureMessage
	}

	return cb
}

func getComponentBuildState(state crv1.ComponentBuildState) types.ComponentBuildState {
	switch state {
	case crv1.ComponentBuildStatePending:
		return constants.ComponentBuildStatePending
	case crv1.ComponentBuildStateQueued:
		return constants.ComponentBuildStateQueued
	case crv1.ComponentBuildStateRunning:
		return constants.ComponentBuildStateRunning
	case crv1.ComponentBuildStateSucceeded:
		return constants.ComponentBuildStateSucceeded
	case crv1.ComponentBuildStateFailed:
		return constants.ComponentBuildStateFailed
	default:
		panic("unreachable")
	}
}

func getComponentBuildFailureMessage(failureInfo crv1.ComponentBuildFailureInfo) string {
	if failureInfo.Internal {
		return "failed due to an internal error"
	}
	return failureInfo.Message
}

func (kb *KubernetesBackend) getPodForComponentBuild(cb *crv1.ComponentBuild) (*corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", kubeconstants.LabelKeyComponentBuildID, cb.Name),
	}
	podsList, err := kb.KubeClientset.CoreV1().Pods(cb.Namespace).List(listOptions)
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
