package app

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
)

func seedRbac(kubeClientset *kubernetes.Clientset) {
	fmt.Println("Seeding rbac...")
	seedRbacComponentBuilder(kubeClientset)
	seedRbacEnvoyXdsApi(kubeClientset)
	seedRbacLatticeControllerManger(kubeClientset)
	seedRbacManagerApi(kubeClientset)
}

var (
	readVerbs          = []string{"get", "watch", "list"}
	readAndCreateVerbs = []string{"get", "watch", "list", "create"}
	readAndDeleteVerbs = []string{"get", "watch", "list", "delete"}
	readAndUpdateVerbs = []string{"get", "watch", "list", "update"}
)

func seedRbacComponentBuilder(kubeClientset *kubernetes.Clientset) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "component-builder",
			Namespace: constants.NamespaceLatticeInternal,
		},
		Rules: []rbacv1.PolicyRule{
			// Read and update lattice component builds
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ComponentBuildResourcePlural},
				Verbs:     readAndUpdateVerbs,
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			Roles(role.Namespace).
			Create(role)
	})

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountComponentBuilder,
			Namespace: role.Namespace,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(sa.Namespace).
			Create(sa)
	})

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "component-builder",
			Namespace: role.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			RoleBindings(rb.Namespace).
			Create(rb)
	})
}

func seedRbacEnvoyXdsApi(kubeClientset *kubernetes.Clientset) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Rules: []rbacv1.PolicyRule{
			// Read kube endpoints
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"endpoints"},
				Verbs:     readVerbs,
			},
			// Read lattice services
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ServiceResourcePlural},
				Verbs:     readVerbs,
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			Roles(role.Namespace).
			Create(role)
	})

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountEnvoyXdsApi,
			Namespace: role.Namespace,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(sa.Namespace).
			Create(sa)
	})

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api",
			Namespace: role.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			RoleBindings(rb.Namespace).
			Create(rb)
	})
}

func seedRbacLatticeControllerManger(kubeClientset *kubernetes.Clientset) {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.MasterNodeComponentLatticeControllerMaster,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice all
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{rbacv1.ResourceAll},
				Verbs:     []string{rbacv1.VerbAll},
			},
			// kube service all
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"services"},
				Verbs:     []string{rbacv1.VerbAll},
			},
			// kube deployment all
			{
				APIGroups: []string{appsv1beta2.GroupName},
				Resources: []string{"deployments"},
				Verbs:     []string{rbacv1.VerbAll},
			},
			// kube job all
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
			Create(clusterRole)
	})

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountLatticeControllerManager,
			Namespace: constants.NamespaceLatticeInternal,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(sa.Namespace).
			Create(sa)
	})

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.MasterNodeComponentLatticeControllerMaster,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(crb)
	})
}

func seedRbacManagerApi(kubeClientset *kubernetes.Clientset) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentManagerApi,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice config read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ConfigResourcePlural},
				Verbs:     readVerbs,
			},
			// lattice system build read and create
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.SystemBuildResourcePlural},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service build read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ComponentBuildResourcePlural},
				Verbs:     readVerbs,
			},
			// lattice component build read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ComponentBuildResourcePlural},
				Verbs:     readVerbs,
			},
			// lattice rollout build and create
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.SystemRolloutResourcePlural},
				Verbs:     readAndCreateVerbs,
			},
			// kube pod read and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods"},
				Verbs:     readAndDeleteVerbs,
			},
			// kube pod/log read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods/log"},
				Verbs:     readVerbs,
			},
			// kube job read
			{
				APIGroups: []string{batchv1.GroupName},
				Resources: []string{"jobs"},
				Verbs:     readVerbs,
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			Roles(role.Namespace).
			Create(role)
	})

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.MasterNodeComponentManagerApi,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice service read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ServiceResourcePlural},
				Verbs:     readVerbs,
			},
			// kube service read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"services"},
				Verbs:     readVerbs,
			},
			// kube node read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"nodes"},
				Verbs:     readVerbs,
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoles().
			Create(clusterRole)
	})

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountManagementApi,
			Namespace: constants.NamespaceLatticeInternal,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			CoreV1().
			ServiceAccounts(sa.Namespace).
			Create(sa)
	})

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentManagerApi,
			Namespace: role.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			RoleBindings(rb.Namespace).
			Create(rb)
	})

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.MasterNodeComponentManagerApi,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			RbacV1().
			ClusterRoleBindings().
			Create(crb)
	})
}
