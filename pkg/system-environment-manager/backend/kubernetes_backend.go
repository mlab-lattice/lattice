package backend

import (
	latticeresource "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	LatticeResourceRestClient rest.Interface
}

func NewKubernetesBackend(kubeconfig string) (*KubernetesBackend, error) {
	// TODO: create in-cluster config if in cluster
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesBackend{
		LatticeResourceRestClient: latticeResourceClient,
	}
	return kb, nil
}
