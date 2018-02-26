package base

import (
	"strconv"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) managerAPIResources(resources *bootstrapper.ClusterResources) {
	internalNamespace := kubeutil.InternalNamespace(b.ClusterID)

	// FIXME: prefix this cluster role with the cluster id so multiple clusters can have different
	// cluster role definitions
	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.MasterNodeComponentManagerAPI,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice system read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralSystem},
				Verbs:     readCreateAndDeleteVerbs,
			},
			// lattice config read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralConfig},
				Verbs:     readVerbs,
			},
			// lattice system build read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralSystemBuild},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralServiceBuild},
				Verbs:     readVerbs,
			},
			// lattice component build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralComponentBuild},
				Verbs:     readVerbs,
			},
			// lattice rollout build and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralSystemRollout},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralService},
				Verbs:     readVerbs,
			},

			// system bootstrapping permissions
			// kube namespace read, update, and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"namespaces"},
				Verbs:     readCreateAndDeleteVerbs,
			},
			// kube service-account read, update, and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"serviceaccounts"},
				Verbs:     readCreateAndDeleteVerbs,
			},
			// kube role-binding read, update, and delete
			{
				APIGroups: []string{rbacv1.GroupName},
				Resources: []string{"rolebindings"},
				Verbs:     readCreateAndDeleteVerbs,
			},
			// kube daemonsets read, update, and delete
			{
				APIGroups: []string{appsv1.GroupName},
				Resources: []string{"daemonsets"},
				Verbs:     readCreateAndDeleteVerbs,
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
	// also need to create component builder SAs for the
	// namespace, so need to have the component builder
	// rules so kube doesn't deny creating the component
	// builder SAs due to privilege escalation
	clusterRole.Rules = append(
		clusterRole.Rules,
		componentBuilderRBACPolicyRules...,
	)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountManagementAPI,
			Namespace: internalNamespace,
		},
	}

	// FIXME: prefix this cluster role binding with the cluster id so multiple clusters can have different
	// cluster role definitions
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
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

	args := []string{
		"--port", strconv.Itoa(int(b.Options.MasterComponents.ManagerAPI.Port)),
		"--cluster-id", string(b.ClusterID),
		"--cloud-provider", b.CloudProviderName,
	}
	args = append(args, b.Options.MasterComponents.ManagerAPI.Args...)
	labels := map[string]string{
		kubeconstants.MasterNodeLabelComponent: kubeconstants.MasterNodeComponentManagerAPI,
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.MasterNodeComponentManagerAPI,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   kubeconstants.MasterNodeComponentManagerAPI,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  kubeconstants.MasterNodeComponentManagerAPI,
							Image: b.Options.MasterComponents.ManagerAPI.Image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      b.Options.MasterComponents.ManagerAPI.Port,
									ContainerPort: b.Options.MasterComponents.ManagerAPI.Port,
								},
							},
						},
					},
					HostNetwork:        b.Options.MasterComponents.ManagerAPI.HostNetwork,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: kubeconstants.ServiceAccountManagementAPI,
					Tolerations: []corev1.Toleration{
						kubeconstants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &kubeconstants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)
	resources.ClusterRoleBindings = append(resources.ClusterRoleBindings, clusterRoleBinding)
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}
