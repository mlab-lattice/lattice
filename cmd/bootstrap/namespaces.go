package main

import (
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

func seedNamespaces(kubeClientset *kubernetes.Clientset) {
	latticeInternalNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.InternalNamespace,
		},
	}

	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.CoreV1().Namespaces().Create(latticeInternalNamespace)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	latticeUserNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(coreconstants.UserSystemNamespace),
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.CoreV1().Namespaces().Create(latticeUserNamespace)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
