package app

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

func seedNamespaces(kubeClientset *kubernetes.Clientset) {
	fmt.Println("Seeding namespaces...")
	namespaces := []*corev1.Namespace{
		// lattice internal namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.NamespaceLatticeInternal,
			},
		},
		// lattice user namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(coreconstants.UserSystemNamespace),
			},
		},
	}
	for _, ns := range namespaces {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.CoreV1().Namespaces().Create(ns)
		})
	}
}
