package main

import (
	"flag"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	localDevDockerRegistry = "lattice-local"
	devDockerRegistry      = "gcr.io/lattice-dev"
)

var (
	kubeconfigPath string
	providerName   string
	userSystemUrl  string
	dev            bool
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&providerName, "provider", "", "name of provider to use")
	flag.StringVar(&userSystemUrl, "user-system-url", "", "url of the user-system definition")
	flag.BoolVar(&dev, "dev", false, "configure to use locally built lattice component docker images")
	flag.Parse()
}

func main() {
	var config *rest.Config
	var err error
	if kubeconfigPath == "" {
		config, err = rest.InClusterConfig()
	} else {
		// TODO: support passing in the context when supported
		// https://github.com/kubernetes/minikube/issues/2100
		//configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}
		configOverrides := &clientcmd.ConfigOverrides{}
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			configOverrides,
		).ClientConfig()
	}

	if err != nil {
		panic(err)
	}

	kubeClientset := clientset.NewForConfigOrDie(config)

	seedNamespaces(kubeClientset)
	seedCrds(config)
	seedRbac(kubeClientset)
	seedConfig(config, userSystemUrl)
	seedEnvoyXdsApi(kubeClientset)
	seedLatticeControllerManager(kubeClientset)
	seedLatticeSystemEnvironmentManagerAPI(kubeClientset)
}

func pollKubeResourceCreation(resourceCreationFunc func() (interface{}, error)) {
	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := resourceCreationFunc()

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
