package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedNamespaces() error {
	fmt.Println("Seeding namespaces")
	namespaces := []*corev1.Namespace{
		// lattice internal namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeutil.GetFullNamespace(b.Options.KubeNamespacePrefix, kubeconstants.NamespaceLatticeInternal),
			},
		},
	}

	for _, namespace := range namespaces {
		_, err := b.KubeClient.CoreV1().Namespaces().Create(namespace)
		if err != nil {
			return err
		}
	}
	return nil
}
