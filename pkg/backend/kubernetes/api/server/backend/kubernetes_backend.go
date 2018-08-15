package backend

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	kubeclientset "k8s.io/client-go/kubernetes"
)

type KubernetesBackend struct {
	namespacePrefix string
	kubeClient      kubeclientset.Interface
	latticeClient   latticeclientset.Interface

	systemBootstrappers []systembootstrapper.Interface
}

func NewKubernetesBackend(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) serverv1.Interface {
	return &KubernetesBackend{
		namespacePrefix: namespacePrefix,
		kubeClient:      kubeClient,
		latticeClient:   latticeClient,
	}
}

func (kb *KubernetesBackend) systemNamespace(systemID v1.SystemID) string {
	return kubeutil.SystemNamespace(kb.namespacePrefix, systemID)
}

func (kb *KubernetesBackend) internalNamespace() string {
	return kubeutil.InternalNamespace(kb.namespacePrefix)
}
