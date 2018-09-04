package backend

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend/system"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	kubeclientset "k8s.io/client-go/kubernetes"
)

type KubernetesBackend struct {
	systems *system.Backend
}

func NewKubernetesBackend(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) *KubernetesBackend {
	return &KubernetesBackend{
		systems: system.NewBackend(namespacePrefix, kubeClient, latticeClient),
	}
}

func (kb *KubernetesBackend) Systems() serverv1.SystemBackend {
	return kb.systems
}
