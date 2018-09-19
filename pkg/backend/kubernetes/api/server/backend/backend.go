package backend

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	backendv1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	kubeclientset "k8s.io/client-go/kubernetes"
)

func NewKubernetesBackend(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) *KubernetesBackend {
	return &KubernetesBackend{
		v1: backendv1.NewBackend(namespacePrefix, kubeClient, latticeClient),
	}
}

type KubernetesBackend struct {
	v1 *backendv1.Backend
}

func (b *KubernetesBackend) V1() serverv1.Interface {
	return b.v1
}
