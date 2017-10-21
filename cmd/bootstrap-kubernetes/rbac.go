package main

import (
	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	batchv1 "k8s.io/api/batch/v1"
	extensions "k8s.io/api/extensions/v1beta1"
)

const (
	kubeEndpointReaderRole   = "kube-endpoint-reader"
	kubeServiceAllRole       = "kube-service-all"
	kubeDeploymentAllRole    = "kube-deployment-all"
	kubeJobAllRole           = "kube-job-all"
	latticeServiceReaderRole = "lattice-service-reader"
	latticeAllRole           = "lattice-all"
)

func seedRbac(kubeClientset *kubernetes.Clientset) {
	seedRbacRoles(kubeClientset)
	seedServiceAccounts(kubeClientset)

	bindEnvoyXdsApiServiceAccountRoles(kubeClientset)
	bindLatticeControllerMangerServiceAccountRoles(kubeClientset)
}

func seedRbacRoles(kubeClientset *kubernetes.Clientset) {
	kubeEndpointReader := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeEndpointReaderRole,
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

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			Roles(string(coreconstants.UserSystemNamespace)).
			Create(kubeEndpointReader)
	})

	latticeServiceReader := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      latticeServiceReaderRole,
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

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			Roles(string(coreconstants.UserSystemNamespace)).
			Create(latticeServiceReader)
	})

	// FIXME: split this up and create individual roles etc for each controller
	latticeResourceAll := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: latticeAllRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{rbacv1.ResourceAll},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoles().
			Create(latticeResourceAll)
	})

	kubeServiceAll := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeServiceAllRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{string(corev1.ResourceServices)},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoles().
			Create(kubeServiceAll)
	})

	kubeDeploymentsAll := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeDeploymentAllRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{extensions.GroupName},
				Resources: []string{"deployments"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoles().
			Create(kubeDeploymentsAll)
	})

	kubeJobAll := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeJobAllRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{batchv1.GroupName},
				Resources: []string{"jobs"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoles().
			Create(kubeJobAll)
	})
}

func seedServiceAccounts(kubeClientset *kubernetes.Clientset) {
	// Create service account for the envoy-xds-api
	envoyXdsApiSa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountEnvoyXdsApi,
			Namespace: string(coreconstants.UserSystemNamespace),
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(string(coreconstants.UserSystemNamespace)).
			Create(envoyXdsApiSa)
	})

	// Create service account for the lattice-controller-manager
	latticeControllerManagerSa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountLatticeControllerManager,
			Namespace: string(constants.NamespaceInternal),
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(string(constants.NamespaceInternal)).
			Create(latticeControllerManagerSa)
	})
}

func bindEnvoyXdsApiServiceAccountRoles(kubeClientset *kubernetes.Clientset) {
	kubeEndpointReaderBind := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api-kube-endpoint-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountEnvoyXdsApi,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     kubeEndpointReaderRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			RoleBindings(string(coreconstants.UserSystemNamespace)).
			Create(kubeEndpointReaderBind)
	})

	latticeServiceReaderBind := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api-lattice-service-reader",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountEnvoyXdsApi,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     latticeServiceReaderRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			RoleBindings(string(coreconstants.UserSystemNamespace)).
			Create(latticeServiceReaderBind)
	})
}

func bindLatticeControllerMangerServiceAccountRoles(kubeClientset *kubernetes.Clientset) {
	latticeAllBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "lattice-controller-manager-lattice-all",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountLatticeControllerManager,
				Namespace: string(constants.NamespaceInternal),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     latticeAllRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(latticeAllBind)
	})

	kubeServiceAllBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "lattice-controller-manager-kube-service-all",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountLatticeControllerManager,
				Namespace: string(constants.NamespaceInternal),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     kubeServiceAllRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(kubeServiceAllBind)
	})

	kubeDeploymentAllBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "lattice-controller-manager-kube-deployment-all",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountLatticeControllerManager,
				Namespace: string(constants.NamespaceInternal),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     kubeDeploymentAllRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(kubeDeploymentAllBind)
	})

	kubeJobAllBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "lattice-controller-manager-kube-job-all",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.ServiceAccountLatticeControllerManager,
				Namespace: string(constants.NamespaceInternal),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     kubeJobAllRole,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(kubeJobAllBind)
	})
}
