package base

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) namespaceResources(resources *bootstrapper.Resources) {
	namespace := &corev1.Namespace{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeutil.InternalNamespace(b.ClusterID),
		},
	}

	resources.Namespaces = append(resources.Namespaces, namespace)
}
