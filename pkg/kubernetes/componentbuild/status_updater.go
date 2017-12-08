package componentbuild

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesStatusUpdater struct {
	LatticeClient latticeclientset.Interface
}

func NewKubernetesStatusUpdater(kubeconfig string) (*KubernetesStatusUpdater, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	latticeClient, err := latticeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesStatusUpdater{
		LatticeClient: latticeClient,
	}
	return kb, nil
}

func (ksu *KubernetesStatusUpdater) UpdateProgress(buildID types.ComponentBuildID, phase types.ComponentBuildPhase) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateProgressInternal(buildID, phase, 1)
}

func (ksu *KubernetesStatusUpdater) updateProgressInternal(buildID types.ComponentBuildID, phase types.ComponentBuildPhase, numRetries int) error {
	cb, err := ksu.LatticeClient.LatticeV1().ComponentBuilds(constants.NamespaceLatticeInternal).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, phase, numRetries-1)
	}

	cb.Status.LastObservedPhase = &phase
	_, err = ksu.LatticeClient.LatticeV1().ComponentBuilds(cb.Namespace).Update(cb)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, phase, numRetries-1)
	}
	return nil
}

func (ksu *KubernetesStatusUpdater) UpdateError(buildID types.ComponentBuildID, internal bool, err error) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateErrorInternal(buildID, internal, err, 1)
}

func (ksu *KubernetesStatusUpdater) updateErrorInternal(buildID types.ComponentBuildID, internal bool, updateErr error, numRetries int) error {
	cb, err := ksu.LatticeClient.LatticeV1().ComponentBuilds(constants.NamespaceLatticeInternal).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateErrorInternal(buildID, internal, updateErr, numRetries-1)
	}

	cb.Status.FailureInfo = &crv1.ComponentBuildFailureInfo{
		Message:  updateErr.Error(),
		Internal: internal,
	}
	_, err = ksu.LatticeClient.LatticeV1().ComponentBuilds(cb.Namespace).Update(cb)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateErrorInternal(buildID, internal, updateErr, numRetries-1)
	}
	return nil
}
