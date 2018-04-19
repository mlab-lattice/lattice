package componentbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"encoding/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesStatusUpdater struct {
	LatticeClient   latticeclientset.Interface
	NamespacePrefix string
}

func NewKubernetesStatusUpdater(namespacePrefix string, kubeconfig string) (*KubernetesStatusUpdater, error) {
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
		LatticeClient:   latticeClient,
		NamespacePrefix: namespacePrefix,
	}
	return kb, nil
}

func (u *KubernetesStatusUpdater) UpdateProgress(buildID v1.ComponentBuildID, systemID v1.SystemID, phase v1.ComponentBuildPhase) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return u.updateProgressInternal(buildID, systemID, phase, 1)
}

func (u *KubernetesStatusUpdater) updateProgressInternal(buildID v1.ComponentBuildID, systemID v1.SystemID, phase v1.ComponentBuildPhase, numRetries int) error {
	namespace := kubeutil.SystemNamespace(u.NamespacePrefix, systemID)
	build, err := u.LatticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return u.updateProgressInternal(buildID, systemID, phase, numRetries-1)
	}

	build.Annotations[constants.AnnotationKeyComponentBuildLastObservedPhase] = string(phase)

	_, err = u.LatticeClient.LatticeV1().ComponentBuilds(build.Namespace).Update(build)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return u.updateProgressInternal(buildID, systemID, phase, numRetries-1)
	}
	return nil
}

func (u *KubernetesStatusUpdater) UpdateError(buildID v1.ComponentBuildID, systemID v1.SystemID, internal bool, err error) error {
	// Retry once since we may lose a race against the controller at the beginning updating the Status.State
	return u.updateErrorInternal(buildID, systemID, internal, err, 1)
}

func (u *KubernetesStatusUpdater) updateErrorInternal(buildID v1.ComponentBuildID, systemID v1.SystemID, internal bool, updateErr error, numRetries int) error {
	namespace := kubeutil.SystemNamespace(u.NamespacePrefix, systemID)
	build, err := u.LatticeClient.LatticeV1().ComponentBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return u.updateErrorInternal(buildID, systemID, internal, updateErr, numRetries-1)
	}

	failureInfo := v1.ComponentBuildFailureInfo{
		Message:  updateErr.Error(),
		Internal: internal,
	}
	data, err := json.Marshal(failureInfo)
	if err != nil {
		return err
	}

	build.Annotations[constants.AnnotationKeyComponentBuildFailureInfo] = string(data)

	_, err = u.LatticeClient.LatticeV1().ComponentBuilds(build.Namespace).Update(build)
	if err != nil {
		if numRetries <= 0 {
			return err
		}
		return u.updateErrorInternal(buildID, systemID, internal, updateErr, numRetries-1)
	}
	return nil
}
