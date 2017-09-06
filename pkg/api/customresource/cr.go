package customresource

import (
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func CreateCustomResourceDefinitions(clientset apiextensionsclient.Interface) ([]*apiextensionsv1beta1.CustomResourceDefinition, error) {
	crds := []*apiextensionsv1beta1.CustomResourceDefinition{}

	for _, topLeveltype := range crv1.TopLevelTypes {
		crdName := topLeveltype.Plural + "." + crv1.GroupName

		crd := &apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: crdName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   crv1.GroupName,
				Version: crv1.SchemeGroupVersion.Version,
				Scope:   topLeveltype.Scope,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Singular: topLeveltype.Singular,
					Plural:   topLeveltype.Plural,
					Kind:     topLeveltype.Kind,
					ListKind: topLeveltype.ListKind,
				},
			},
		}

		_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if err != nil {
			return nil, err
		}

		// wait for CRD being established
		err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
			crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextensionsv1beta1.Established:
					if cond.Status == apiextensionsv1beta1.ConditionTrue {
						return true, err
					}
				case apiextensionsv1beta1.NamesAccepted:
					if cond.Status == apiextensionsv1beta1.ConditionFalse {
						fmt.Printf("Name conflict: %v\n", cond.Reason)
					}
				}
			}

			return false, err
		})

		if err != nil {
			deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crdName, nil)
			if deleteErr != nil {
				return nil, errors.NewAggregate([]error{err, deleteErr})
			}
			return nil, err
		}

		crds = append(crds, crd)
	}

	return crds, nil
}
