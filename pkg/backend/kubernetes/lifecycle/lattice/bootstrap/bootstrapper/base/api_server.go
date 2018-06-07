package base

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) apiServerResources(resources *bootstrapper.Resources) {
	internalNamespace := kubeutil.InternalNamespace(b.NamespacePrefix)
	name := fmt.Sprintf("%v-%v", b.NamespacePrefix, constants.ControlPlaneServiceAPIServer)

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
			// lattice system read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralSystem},
				Verbs:     ReadCreateAndDeleteVerbs,
			},
			// lattice config read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralConfig},
				Verbs:     ReadVerbs,
			},
			// lattice system build read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralBuild},
				Verbs:     ReadAndCreateVerbs,
			},
			// lattice service build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralServiceBuild},
				Verbs:     ReadVerbs,
			},
			// lattice component build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralContainerBuild},
				Verbs:     ReadVerbs,
			},
			// lattice deploy read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralDeploy},
				Verbs:     ReadAndCreateVerbs,
			},
			// lattice teardown read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralTeardown},
				Verbs:     ReadAndCreateVerbs,
			},
			// lattice service read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralService},
				Verbs:     ReadVerbs,
			},
			// lattice node pool read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralNodePool},
				Verbs:     ReadVerbs,
			},

			// kube pod read and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods"},
				Verbs:     ReadAndDeleteVerbs,
			},
			// kube pod/log read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods/log"},
				Verbs:     ReadVerbs,
			},
			// kube job read
			{
				APIGroups: []string{batchv1.GroupName},
				Resources: []string{"jobs"},
				Verbs:     ReadVerbs,
			},
			// kube service read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"services"},
				Verbs:     ReadVerbs,
			},
			// kube node read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"nodes"},
				Verbs:     ReadVerbs,
			},
			// kube secret
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secrets"},
				Verbs:     ReadCreateUpdateAndDeleteVerbs,
			},
		},
	}
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ServiceAccountAPIServer,
			Namespace: internalNamespace,
		},
	}
	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
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
		"--port", strconv.Itoa(int(b.Options.MasterComponents.APIServer.Port)),
		"--namespace-prefix", b.NamespacePrefix,
		"--alsologtostderr",
	}
	args = append(args, b.Options.MasterComponents.APIServer.Args...)
	labels := map[string]string{
		constants.LabelKeyControlPlaneService: constants.ControlPlaneServiceAPIServer,
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ControlPlaneServiceAPIServer,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.ControlPlaneServiceAPIServer,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.ControlPlaneServiceAPIServer,
							Image: b.Options.MasterComponents.APIServer.Image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      b.Options.MasterComponents.APIServer.Port,
									ContainerPort: b.Options.MasterComponents.APIServer.Port,
								},
							},
						},
					},
					HostNetwork:        b.Options.MasterComponents.APIServer.HostNetwork,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountAPIServer,
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
