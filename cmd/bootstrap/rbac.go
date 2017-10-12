package main

import (
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func seedRbac(kubeClientset *kubernetes.Clientset) {
	// Create RBAC roles
	kubeEndpointReader := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-endpoint-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}

	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.
			RbacV1().
			Roles(string(coreconstants.UserSystemNamespace)).
			Create(kubeEndpointReader)

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	latticeServiceReader := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lattice-service-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ServiceResourcePlural},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.
			RbacV1().
			Roles(string(coreconstants.UserSystemNamespace)).
			Create(latticeServiceReader)

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	// Create service account for the envoy-xds-api
	envoyXdsApiSa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountEnvoyXdsApi,
			Namespace: string(coreconstants.UserSystemNamespace),
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.
			CoreV1().
			ServiceAccounts(string(coreconstants.UserSystemNamespace)).
			Create(envoyXdsApiSa)

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	// Bind kube-endpoint-reader and lattice-service-reader to the envoy-xds-api service account
	kubeEndpointReaderBind := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api-kube-endpoint-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      envoyXdsApiSa.Name,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     kubeEndpointReader.Name,
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.
			RbacV1().
			RoleBindings(string(coreconstants.UserSystemNamespace)).
			Create(kubeEndpointReaderBind)

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	latticeServiceReaderBind := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api-lattice-service-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      envoyXdsApiSa.Name,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     latticeServiceReader.Name,
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.
			RbacV1().
			RoleBindings(string(coreconstants.UserSystemNamespace)).
			Create(latticeServiceReaderBind)

		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
