package base

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/util"
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
			Name: kubeutil.InternalNamespace(b.ClusterID),
		},
	}

	if b.Options.DryRun {
		return []interface{}{namespace}, nil
	}

	fmt.Println("Seeding namespaces")

	result, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.CoreV1().Namespaces().Create(namespace)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{result}, nil
}
