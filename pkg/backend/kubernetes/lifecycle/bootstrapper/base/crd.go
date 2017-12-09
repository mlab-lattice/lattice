package base

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func (b *DefaultBootstrapper) seedCRD() error {
	fmt.Println("Seeding CRDs")

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(b.KubeConfig)
	if err != nil {
		return err
	}
	_, err = customresource.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		return err
	}
	return nil
}
