package main

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

const (
	kubeEndpointReaderRole           = "kube-endpoint-reader"
	kubeServiceAllRole               = "kube-service-all"
	kubeServiceReadRole              = "kub-service-read"
	kubeDeploymentAllRole            = "kube-deployment-all"
	kubeJobAllRole                   = "kube-job-all"
	latticeServiceReaderRole         = "lattice-service-reader"
	latticeAllRole                   = "lattice-all"
	latticeConfigReadRole            = "lattice-config-read"
	latticeBuildsReadAndCreateRole   = "lattice-builds-read-and-create"
	latticeRolloutsReadAndCreateRole = "lattice-rollouts-read-and-create"
)

func seedRbac(kubeClientset *kubernetes.Clientset) {
	fmt.Println("Seeding rbac...")
	seedRbacRoles(kubeClientset)
	seedServiceAccounts(kubeClientset)

	bindEnvoyXdsApiServiceAccountRoles(kubeClientset)
	bindLatticeControllerMangerServiceAccountRoles(kubeClientset)
	bindLatticeSystemEnvironmentMangerApiServiceAccountRoles(kubeClientset)
}

var readVerbs []string = []string{"get", "watch", "list"}
var readAndCreateVerbs []string = []string{"get", "watch", "list", "create"}

func seedRbacRoles(kubeClientset *kubernetes.Clientset) {
	roles := []*rbacv1.Role{
		// kube Endpoint reader
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeEndpointReaderRole,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"endpoints"},
					Verbs:     readVerbs,
				},
			},
		},
		// lattice Builds read and create
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      latticeBuildsReadAndCreateRole,
				Namespace: constants.NamespaceLatticeInternal,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{crv1.GroupName},
					Resources: []string{crv1.SystemBuildResourcePlural},
					Verbs:     readAndCreateVerbs,
				},
			},
		},
		// lattice Rollouts read and create
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      latticeRolloutsReadAndCreateRole,
				Namespace: constants.NamespaceLatticeInternal,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{crv1.GroupName},
					Resources: []string{crv1.SystemRolloutResourcePlural},
					Verbs:     readAndCreateVerbs,
				},
			},
		},
		// lattice Config read
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      latticeConfigReadRole,
				Namespace: constants.NamespaceLatticeInternal,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{crv1.GroupName},
					Resources: []string{crv1.ConfigResourcePlural},
					Verbs:     readVerbs,
				},
			},
		},
	}

	for _, r := range roles {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				Roles(r.Namespace).
				Create(r)
		})
	}

	clusterRoles := []*rbacv1.ClusterRole{
		// lattice resources all
		// FIXME: split this up and create individual roles etc for each controller
		//        need to figure out how to distribute these to each controller's client
		{
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
		},
		// lattice Service reader
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: latticeServiceReaderRole,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{crv1.GroupName},
					Resources: []string{crv1.ServiceResourcePlural},
					Verbs:     readVerbs,
				},
			},
		},
		// kube Service reader
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeServiceReadRole,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{string(corev1.ResourceServices)},
					Verbs:     readVerbs,
				},
			},
		},
		// kube Services all
		{
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
		},
		// kube Deployments all
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeDeploymentAllRole,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{appsv1beta2.GroupName},
					Resources: []string{"deployments"},
					Verbs:     []string{rbacv1.VerbAll},
				},
			},
		},
		// kube Jobs all
		{
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
		},
	}

	for _, cr := range clusterRoles {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				ClusterRoles().
				Create(cr)
		})
	}

}

func seedServiceAccounts(kubeClientset *kubernetes.Clientset) {
	serviceAccounts := []*corev1.ServiceAccount{
		// envoy-xds-api
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ServiceAccountEnvoyXdsApi,
				Namespace: string(coreconstants.UserSystemNamespace),
			},
		},
		// lattice-controller-manager
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ServiceAccountLatticeControllerManager,
				Namespace: constants.NamespaceLatticeInternal,
			},
		},
		// lattice-system-environment-manager-api

		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
				Namespace: constants.NamespaceLatticeInternal,
			},
		},
	}

	for _, sa := range serviceAccounts {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				CoreV1().
				ServiceAccounts(sa.Namespace).
				Create(sa)
		})
	}
}

func bindEnvoyXdsApiServiceAccountRoles(kubeClientset *kubernetes.Clientset) {
	roleBindings := []*rbacv1.RoleBinding{
		// kube endpoint reader
		{
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
		},
		// lattice Service reader
		{
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
				Kind:     "ClusterRole",
				Name:     latticeServiceReaderRole,
			},
		},
	}

	for _, rb := range roleBindings {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				RoleBindings(rb.Namespace).
				Create(rb)
		})
	}
}

func bindLatticeControllerMangerServiceAccountRoles(kubeClientset *kubernetes.Clientset) {
	clusterRoleBindings := []*rbacv1.ClusterRoleBinding{
		// lattice all
		// FIXME: split this up and create individual roles etc for each controller.
		//        need to figure out how to distribute these to each controller's client
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-controller-manager-lattice-all",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeControllerManager,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     latticeAllRole,
			},
		},
		// kube Service all
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-controller-manager-kube-service-all",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeControllerManager,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     kubeServiceAllRole,
			},
		},
		// kube Deployment all
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-controller-manager-kube-deployment-all",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeControllerManager,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     kubeDeploymentAllRole,
			},
		},
		// kube Job all
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-controller-manager-kube-job-all",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeControllerManager,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     kubeJobAllRole,
			},
		},
	}

	for _, crb := range clusterRoleBindings {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				ClusterRoleBindings().
				Create(crb)
		})
	}
}

func bindLatticeSystemEnvironmentMangerApiServiceAccountRoles(kubeClientset *kubernetes.Clientset) {
	roleBindings := []*rbacv1.RoleBinding{
		// lattice Configs read
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lattice-system-environment-manager-api-configs-read",
				Namespace: constants.NamespaceLatticeInternal,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     latticeConfigReadRole,
			},
		},
		// lattice Builds read and create
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lattice-system-environment-manager-api-builds-read-and-create",
				Namespace: constants.NamespaceLatticeInternal,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     latticeBuildsReadAndCreateRole,
			},
		},
		// lattice Rollouts read and create
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lattice-system-environment-manager-api-rollouts-read-and-create",
				Namespace: constants.NamespaceLatticeInternal,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     latticeRolloutsReadAndCreateRole,
			},
		},
	}

	for _, rb := range roleBindings {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				RoleBindings(rb.Namespace).
				Create(rb)
		})
	}

	clusterRoleBindings := []*rbacv1.ClusterRoleBinding{
		// kube Service read
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-system-environment-manager-api-kube-service-read",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     kubeServiceReadRole,
			},
		},
		// lattice Service read
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lattice-system-environment-manager-api-lattice-service-read",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					Namespace: constants.NamespaceLatticeInternal,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     latticeServiceReaderRole,
			},
		},
	}

	for _, crb := range clusterRoleBindings {
		pollKubeResourceCreation(func() (interface{}, error) {
			return kubeClientset.
				RbacV1().
				ClusterRoleBindings().
				Create(crb)
		})
	}
}
