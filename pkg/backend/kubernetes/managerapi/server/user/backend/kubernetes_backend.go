package backend

import (
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	"github.com/mlab-lattice/system/pkg/types"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	ClusterID     types.ClusterID
	KubeClient    kubeclientset.Interface
	LatticeClient latticeclientset.Interface
}

func NewKubernetesBackend(clusterID types.ClusterID, kubeconfig string) (*KubernetesBackend, error) {
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
		ClusterID:     clusterID,
		KubeClient:    kubeClient,
		LatticeClient: latticeClient,
	}
	return kb, nil
}
