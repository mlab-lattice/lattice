package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedNamespaces() ([]interface{}, error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, kubeconstants.NamespaceLatticeInternal),
		},
	}

	if b.Options.DryRun {
		return []interface{}{namespace}, nil
	}

	fmt.Println("Seeding namespaces")

	namespace, err := b.KubeClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		return nil, err
	}
	return []interface{}{namespace}, nil
}
