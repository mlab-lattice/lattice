package componentbuilder

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"encoding/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesStatusUpdater struct {
	LatticeClient latticeclientset.Interface
	ClusterID     types.LatticeID
}

func NewKubernetesStatusUpdater(clusterID types.LatticeID, kubeconfig string) (*KubernetesStatusUpdater, error) {
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
		ClusterID:     clusterID,
	}
	return kb, nil
}

func (ksu *KubernetesStatusUpdater) UpdateProgress(buildID types.ComponentBuildID, systemID types.SystemID, phase types.ComponentBuildPhase) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateProgressInternal(buildID, systemID, phase, 1)
}

func (ksu *KubernetesStatusUpdater) updateProgressInternal(buildID types.ComponentBuildID, systemID types.SystemID, phase types.ComponentBuildPhase, numRetries int) error {
	namespace := kubeutil.SystemNamespace(ksu.ClusterID, systemID)
	build, err := ksu.LatticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, systemID, phase, numRetries-1)
	}

	build.Annotations[constants.AnnotationKeyComponentBuildLastObservedPhase] = string(phase)

	_, err = ksu.LatticeClient.LatticeV1().ComponentBuilds(build.Namespace).Update(build)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateProgressInternal(buildID, systemID, phase, numRetries-1)
	}
	return nil
}

func (ksu *KubernetesStatusUpdater) UpdateError(buildID types.ComponentBuildID, systemID types.SystemID, internal bool, err error) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return ksu.updateErrorInternal(buildID, systemID, internal, err, 1)
}

func (ksu *KubernetesStatusUpdater) updateErrorInternal(buildID types.ComponentBuildID, systemID types.SystemID, internal bool, updateErr error, numRetries int) error {
	namespace := kubeutil.SystemNamespace(ksu.ClusterID, systemID)
	build, err := ksu.LatticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateErrorInternal(buildID, systemID, internal, updateErr, numRetries-1)
	}

	failureInfo := types.ComponentBuildFailureInfo{
		Message:  updateErr.Error(),
		Internal: internal,
	}
	data, err := json.Marshal(failureInfo)
	if err != nil {
		return err
	}

	build.Annotations[constants.AnnotationKeyComponentBuildFailureInfo] = string(data)

	_, err = ksu.LatticeClient.LatticeV1().ComponentBuilds(build.Namespace).Update(build)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return ksu.updateErrorInternal(buildID, systemID, internal, updateErr, numRetries-1)
	}
	return nil
}
