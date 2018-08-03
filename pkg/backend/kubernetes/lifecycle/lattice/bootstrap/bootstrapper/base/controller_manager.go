package base

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) controllerManagerResources(resources *bootstrapper.Resources) {
	internalNamespace := kubeutil.InternalNamespace(b.NamespacePrefix)
	name := fmt.Sprintf("%v-%v", b.NamespacePrefix, constants.ControlPlaneServiceLatticeControllerManager)

	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice all
			{
				APIGroups: []string{latticev1.GroupName},
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
			// kube pod read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods"},
				Verbs:     ReadVerbs,
			},
			// kube node read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"nodes"},
				Verbs:     ReadVerbs,
			},

			// system bootstrapping permissions
			// kube namespace read, update, and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"namespaces"},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
			// kube service-account read, update, and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"serviceaccounts"},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
			// kube role-binding read, update, and delete
			{
				APIGroups: []string{rbacv1.GroupName},
				Resources: []string{"rolebindings"},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
			// kube daemonsets read, update, and delete
			{
				APIGroups: []string{appsv1.GroupName},
				Resources: []string{"daemonsets"},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
			// kube secret create
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secrets"},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
		},
	}
	// also need to create component builder SAs for the system namespace when bootstrapping a system,
	// so need to have the component builder rules so kube doesn't deny creating the component
	// builder SAs due to privilege escalation
	clusterRole.Rules = append(
		clusterRole.Rules,
		containerBuilderRBACPolicyRules...,
	)
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ControlPlaneServiceLatticeControllerManager,
			Namespace: internalNamespace,
		},
	}
	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}
	resources.ClusterRoleBindings = append(resources.ClusterRoleBindings, clusterRoleBinding)

	args := []string{
		"--cloud-provider", b.CloudProviderName,
		"--lattice-id", string(b.LatticeID),
		"--namespace-prefix", b.NamespacePrefix,
		"--internal-dns-domain", b.InternalDNSDomain,
		"--alsologtostderr",
	}
	args = append(args, b.Options.MasterComponents.LatticeControllerManager.Args...)

	labels := map[string]string{
		constants.LabelKeyControlPlaneService: constants.ControlPlaneServiceLatticeControllerManager,
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ControlPlaneServiceLatticeControllerManager,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.ControlPlaneServiceLatticeControllerManager,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.ControlPlaneServiceLatticeControllerManager,
							Image: b.Options.MasterComponents.LatticeControllerManager.Image,
							Args:  args,
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationKubernetesMasterNode,
						constants.TolerationLatticeMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}
