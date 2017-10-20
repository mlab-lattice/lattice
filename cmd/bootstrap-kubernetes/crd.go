package main

import (
	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/client-go/rest"
)

func seedCrds(kubeconfig *rest.Config) {
	apiextensionsclientset, err := apiextensionsclient.NewForConfig(kubeconfig)
	if err != nil {
		panic(err)
	}
	_, err = crdclient.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		panic(err)
	}
}
