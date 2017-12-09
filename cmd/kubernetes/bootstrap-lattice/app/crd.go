package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func seedCrds() {
	fmt.Println("Seeding CRDs...")

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}
	_, err = customresource.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		panic(err)
	}
}
