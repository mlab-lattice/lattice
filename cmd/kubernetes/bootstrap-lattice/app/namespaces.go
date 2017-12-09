package app

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func seedNamespaces() {
	fmt.Println("Seeding namespaces...")
	namespaces := []*corev1.Namespace{
		// lattice internal namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeconstants.NamespaceLatticeInternal,
			},
		},
		// lattice user namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(constants.UserSystemNamespace),
			},
		},
	}
	for _, ns := range namespaces {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClient.CoreV1().Namespaces().Create(ns)
		})
	}
}
