package base

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func (b *DefaultBootstrapper) seedCRD() ([]interface{}, error) {
	if b.Options.DryRun {
		crds := customresource.GetCustomResourceDefinitions()
		return convertCRDsToInterface(crds), nil
	}

	fmt.Println("Seeding CRDs")

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(b.KubeConfig)
	if err != nil {
		return nil, err
	}
	crds, err := customresource.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		return nil, err
	}

	return convertCRDsToInterface(crds), nil
}

func convertCRDsToInterface(crds []*apiextensionsv1beta1.CustomResourceDefinition) []interface{} {
	var interfaces []interface{}
	for _, crd := range crds {
		interfaces = append(interfaces, interface{}(crd))
	}

	return interfaces
}
