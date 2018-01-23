package customresource

import (
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/rest"
)

func CreateCustomResourceDefinitions(
	definitions []*apiextensionsv1beta1.CustomResourceDefinition,
	kubeConfig *rest.Config,
) ([]*apiextensionsv1beta1.CustomResourceDefinition, error) {
	client, err := apiextensionsclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	var definitionResults []*apiextensionsv1beta1.CustomResourceDefinition
	for _, definition := range definitions {
		_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(definition)
		if errors.IsAlreadyExists(err) {
			fmt.Printf("CRD %v already exists", definition.Name)
		}

		// wait for CRD being established
		err = wait.Poll(500*time.Millisecond, 20*time.Second, func() (bool, error) {
			definition, err = client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(definition.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			for _, cond := range definition.Status.Conditions {
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
			return nil, err
		}

		definitionResults = append(definitionResults, definition)
	}

	return definitionResults, nil
}
