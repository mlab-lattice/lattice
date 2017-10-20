package main

import (
	"flag"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	localDevDockerRegistry = "lattice-local-dev"
	devDockerRegistry      = "gcr.io/lattice-dev"
)

var (
	kubeconfigPath string
	providerName   string
	dev            bool
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&providerName, "provider", "", "path to kubeconfig file")
	flag.BoolVar(&dev, "dev", false, "configure to use locally built lattice component docker images")
	flag.Parse()
}

func main() {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	kubeClientset := clientset.NewForConfigOrDie(kubeconfig)

	seedNamespaces(kubeClientset)
	seedCrds(kubeconfig)
	seedRbac(kubeClientset)
	seedConfig(kubeconfig)
	seedEnvoyXdsApi(kubeClientset)
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
