package main

import (
	"flag"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfigPath string
	providerName   string
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&providerName, "provider", "", "path to kubeconfig file")
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
