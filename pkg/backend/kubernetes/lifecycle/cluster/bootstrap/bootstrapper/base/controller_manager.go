package base

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"github.com/mlab-lattice/system/pkg/terraform"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) controllerManagerResources(resources *bootstrapper.ClusterResources) {
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
			Name: kubeconstants.MasterNodeComponentLatticeControllerManager,
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
		},
	}
	// also need to create component builder SAs for the system namespace when bootstrapping a system,
	// so need to have the component builder rules so kube doesn't deny creating the component
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
			Name:      kubeconstants.ServiceAccountLatticeControllerManager,
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
			Name: kubeconstants.MasterNodeComponentLatticeControllerManager,
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

	args := []string{"--cloud-provider", b.CloudProviderName, "--cluster-id", string(b.ClusterID)}
	args = append(args, b.Options.MasterComponents.LatticeControllerManager.Args...)

	if b.Options.MasterComponents.LatticeControllerManager.TerraformModulePath != "" {
		args = append(
			args,
			"--terraform-module-path", b.Options.MasterComponents.LatticeControllerManager.TerraformModulePath,
		)
	}

	if b.Options.TerraformOptions.Backend.S3 != nil {
		args = append(
			args,
			"--terraform-backend", terraform.BackendS3,
			"--terraform-backend-var", fmt.Sprintf("bucket=%v", b.Options.TerraformOptions.Backend.S3.Bucket),
		)
	}

	labels := map[string]string{
		kubeconstants.MasterNodeLabelComponent: kubeconstants.MasterNodeComponentLatticeControllerManager,
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.MasterNodeComponentLatticeControllerManager,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   kubeconstants.MasterNodeComponentLatticeControllerManager,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  kubeconstants.MasterNodeComponentLatticeControllerManager,
							Image: b.Options.MasterComponents.LatticeControllerManager.Image,
							Args:  args,
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: kubeconstants.ServiceAccountLatticeControllerManager,
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
