package componentbuild

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	latticeresource "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesStatusUpdater struct {
	LatticeResourceClient rest.Interface
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

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesStatusUpdater{
		LatticeResourceClient: latticeResourceClient,
	}
	return kb, nil
}

func (ksu *KubernetesStatusUpdater) UpdateProgress(buildID coretypes.ComponentBuildID, phase coretypes.ComponentBuildPhase) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateProgressInternal(buildID, phase, 1)
}

func (ksu *KubernetesStatusUpdater) updateProgressInternal(buildID coretypes.ComponentBuildID, phase coretypes.ComponentBuildPhase, numRetries int) error {
	cb := &crv1.ComponentBuild{}
	err := ksu.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(string(buildID)).
		Do().
		Into(cb)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, phase, numRetries-1)
	}

	cb.Status.LastObservedPhase = &phase
	err = ksu.LatticeResourceClient.Put().
		Namespace(cb.Namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cb.Name).
		Body(cb).
		Do().
		Into(nil)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, phase, numRetries-1)
	}
	return nil
}

func (ksu *KubernetesStatusUpdater) UpdateError(buildID coretypes.ComponentBuildID, internal bool, err error) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateErrorInternal(buildID, internal, err, 1)
}

func (ksu *KubernetesStatusUpdater) updateErrorInternal(buildID coretypes.ComponentBuildID, internal bool, updateErr error, numRetries int) error {
	cb := &crv1.ComponentBuild{}
	err := ksu.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(string(buildID)).
		Do().
		Into(cb)
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
	err = ksu.LatticeResourceClient.Put().
		Namespace(cb.Namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cb.Name).
		Body(cb).
		Do().
		Into(nil)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateErrorInternal(buildID, internal, updateErr, numRetries-1)
	}
	return nil
}
