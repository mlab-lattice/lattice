package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/util"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

var (
	readVerbs          = []string{"get", "watch", "list"}
	readAndCreateVerbs = []string{"get", "watch", "list", "create"}
	readAndDeleteVerbs = []string{"get", "watch", "list", "delete"}
	readAndUpdateVerbs = []string{"get", "watch", "list", "update"}
) /**/

func (b *DefaultBootstrapper) seedRBAC() ([]interface{}, error) {
	if !b.Options.DryRun {
		fmt.Println("Seeding rbac")
	}

	rbacSeedFuncs := []func() ([]interface{}, error){
		b.seedRBACComponentBuilder,
		b.seedRBACEnvoyXDSAPI,
		b.seedRBACLatticeControllerManger,
		b.seedRBACManagerAPI,
	}

	objects := []interface{}{}
	for _, rbacSeedFunc := range rbacSeedFuncs {
		additionalObjects, err := rbacSeedFunc()
		if err != nil {
			return nil, err
		}
		objects = append(objects, additionalObjects...)
	}
	return objects, nil
}

func (b *DefaultBootstrapper) seedRBACComponentBuilder() ([]interface{}, error) {
	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, kubeconstants.NamespaceLatticeInternal)
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountComponentBuilder,
			Namespace: role.Namespace,
		},
	}

	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	if b.Options.DryRun {
		return []interface{}{role, sa, rb}, nil
	}

	roleResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().Roles(role.Namespace).Create(role)
	})
	if err != nil {
		return nil, err
	}

	saResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	})
	if err != nil {
		return nil, err
	}

	rbResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().RoleBindings(rb.Namespace).Create(rb)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{roleResult, saResult, rbResult}, nil
}

func (b *DefaultBootstrapper) seedRBACEnvoyXDSAPI() ([]interface{}, error) {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	if b.Options.DryRun {
		return []interface{}{clusterRole}, nil
	}

	clusterRoleResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{clusterRoleResult}, nil
}

func (b *DefaultBootstrapper) seedRBACLatticeControllerManger() ([]interface{}, error) {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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
				APIGroups: []string{appsv1.GroupName},
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

	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, kubeconstants.NamespaceLatticeInternal)

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountLatticeControllerManager,
			Namespace: namespace,
		},
	}

	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	if b.Options.DryRun {
		return []interface{}{clusterRole, sa, crb}, nil
	}

	clusterRoleResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole)
	})
	if err != nil {
		return nil, err
	}

	saResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	})
	if err != nil {
		return nil, err
	}

	crbResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().ClusterRoleBindings().Create(crb)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{clusterRoleResult, saResult, crbResult}, nil
}

func (b *DefaultBootstrapper) seedRBACManagerAPI() ([]interface{}, error) {
	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, kubeconstants.NamespaceLatticeInternal)
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountManagementAPI,
			Namespace: role.Namespace,
		},
	}

	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
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

	if b.Options.DryRun {
		return []interface{}{role, clusterRole, sa, rb, crb}, nil
	}

	roleResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().Roles(role.Namespace).Create(role)
	})
	if err != nil {
		return nil, err
	}

	clusterRoleResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().ClusterRoles().Create(clusterRole)
	})
	if err != nil {
		return nil, err
	}

	saResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	})
	if err != nil {
		return nil, err
	}

	rbResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().RoleBindings(rb.Namespace).Create(rb)
	})
	if err != nil {
		return nil, err
	}

	crbResult, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.RbacV1().ClusterRoleBindings().Create(crb)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{roleResult, clusterRoleResult, saResult, rbResult, crbResult}, nil
}
