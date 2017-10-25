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
	namespaces := []*corev1.Namespace{
		// lattice internal namespace
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.NamespaceInternal,
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
