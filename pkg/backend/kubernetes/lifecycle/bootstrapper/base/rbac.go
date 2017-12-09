package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
)

var (
	readVerbs          = []string{"get", "watch", "list"}
	readAndCreateVerbs = []string{"get", "watch", "list", "create"}
	readAndDeleteVerbs = []string{"get", "watch", "list", "delete"}
	readAndUpdateVerbs = []string{"get", "watch", "list", "update"}
) /**/

func (b *DefaultBootstrapper) seedRBAC() error {
	fmt.Println("Seeding rbac")

	rbacSeedFuncs := []func() error{
		b.seedRBACComponentBuilder,
		b.seedRBACEnvoyXDSAPI,
		b.seedRBACLatticeControllerManger,
		b.seedRBACManagerAPI,
	}

	for _, rbacSeedFunc := range rbacSeedFuncs {
		if err := rbacSeedFunc(); err != nil {
			return err
		}
	}
	return nil
}

func (b *DefaultBootstrapper) seedRBACComponentBuilder() error {
	namespace := kubeutil.GetFullNamespace(b.Options.KubeNamespacePrefix, kubeconstants.NamespaceLatticeInternal)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.InternalComponentComponentBuilder,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			// Read and update lattice component builds
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralComponentBuild},
				Verbs:     readAndUpdateVerbs,
			},
		},
	}

	if _, err := b.KubeClient.RbacV1().Roles(role.Namespace).Create(role); err != nil {
		return err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountComponentBuilder,
			Namespace: role.Namespace,
		},
	}

	if _, err := b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa); err != nil {
		return err
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.InternalComponentComponentBuilder,
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

	if _, err := b.KubeClient.RbacV1().RoleBindings(rb.Namespace).Create(rb); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBootstrapper) seedRBACEnvoyXDSAPI() error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.InternalComponentEnvoyXDSAPI,
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
				Resources: []string{crv1.ResourcePluralService},
				Verbs:     readVerbs,
			},
		},
	}

	if _, err := b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBootstrapper) seedRBACLatticeControllerManger() error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.MasterNodeComponentLatticeControllerManager,
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

	if _, err := b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole); err != nil {
		return err
	}

	namespace := kubeutil.GetFullNamespace(b.Options.KubeNamespacePrefix, kubeconstants.NamespaceLatticeInternal)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountLatticeControllerManager,
			Namespace: namespace,
		},
	}

	if _, err := b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa); err != nil {
		return err
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.MasterNodeComponentLatticeControllerManager,
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

	if _, err := b.KubeClient.RbacV1().ClusterRoleBindings().Create(crb); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBootstrapper) seedRBACManagerAPI() error {
	namespace := kubeutil.GetFullNamespace(b.Options.KubeNamespacePrefix, kubeconstants.NamespaceLatticeInternal)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.MasterNodeComponentManagerAPI,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice config read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralConfig},
				Verbs:     readVerbs,
			},
			// lattice system build read and create
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralSystemBuild},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service build read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralServiceBuild},
				Verbs:     readVerbs,
			},
			// lattice component build read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralComponentBuild},
				Verbs:     readVerbs,
			},
			// lattice rollout build and create
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralSystemRollout},
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

	if _, err := b.KubeClient.RbacV1().Roles(role.Namespace).Create(role); err != nil {
		return err
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.MasterNodeComponentManagerAPI,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice service read
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralService},
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

	if _, err := b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole); err != nil {
		return err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountManagementAPI,
			Namespace: role.Namespace,
		},
	}

	if _, err := b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa); err != nil {
		return err
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.MasterNodeComponentManagerAPI,
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

	if _, err := b.KubeClient.RbacV1().RoleBindings(rb.Namespace).Create(rb); err != nil {
		return err
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.MasterNodeComponentManagerAPI,
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

	if _, err := b.KubeClient.RbacV1().ClusterRoleBindings().Create(crb); err != nil {
		return err
	}
	return nil
}
