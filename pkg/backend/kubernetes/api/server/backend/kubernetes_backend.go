package backend

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	namespacePrefix string
	kubeClient      kubeclientset.Interface
	latticeClient   latticeclientset.Interface

	systemBootstrappers []systembootstrapper.Interface
}

func NewKubernetesBackend(
	namespacePrefix string,
	kubeconfig string,
) (*KubernetesBackend, error) {
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

	kubeClient, err := kubeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	latticeClient, err := latticeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesBackend{
		namespacePrefix: namespacePrefix,
		kubeClient:      kubeClient,
		latticeClient:   latticeClient,
	}
	return kb, nil
}

func (kb *KubernetesBackend) systemNamespace(systemID v1.SystemID) string {
	return kubeutil.SystemNamespace(kb.namespacePrefix, systemID)
}

func (kb *KubernetesBackend) internalNamespace() string {
	return kubeutil.InternalNamespace(kb.namespacePrefix)
}
