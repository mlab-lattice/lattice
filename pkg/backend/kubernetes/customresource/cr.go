package customresource

import (
	"fmt"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"
)

func CreateCustomResourceDefinitions(clientset apiextensionsclient.Interface) ([]*apiextensionsv1beta1.CustomResourceDefinition, error) {
	crds := GetCustomResourceDefinitions()

	for _, crd := range crds {
		_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				glog.Warningf("CRD %v already exists", crd.Name)
				continue
			}

			return nil, err
		}

		// wait for CRD being established
		err = wait.Poll(500*time.Millisecond, 20*time.Second, func() (bool, error) {
			crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
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
			deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, nil)
			if deleteErr != nil {
				return nil, errors.NewAggregate([]error{err, deleteErr})
			}
			return nil, err
		}

		crds = append(crds, crd)
	}

	return crds, nil
}

func GetCustomResourceDefinitions() []*apiextensionsv1beta1.CustomResourceDefinition {
	crds := []*apiextensionsv1beta1.CustomResourceDefinition{}
	for _, resource := range crv1.Resources {
		crdName := resource.Plural + "." + crv1.GroupName

		crd := &apiextensionsv1beta1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: apiextensionsv1beta1.GroupName + "/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: crdName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   crv1.GroupName,
				Version: crv1.SchemeGroupVersion.Version,
				Scope:   resource.Scope,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Singular:   resource.Singular,
					Plural:     resource.Plural,
					ShortNames: resource.ShortNames,
					Kind:       resource.Kind,
					ListKind:   resource.ListKind,
				},
			},
		}
		crds = append(crds, crd)
	}
	return crds
}