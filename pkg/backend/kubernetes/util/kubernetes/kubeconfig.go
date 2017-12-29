package kubernetes

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewConfig(kubeconfigPath, kubeContext string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		return rest.InClusterConfig()
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides,
	).ClientConfig()
}
