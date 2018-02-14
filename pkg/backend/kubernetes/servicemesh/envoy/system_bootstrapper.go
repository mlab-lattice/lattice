package envoy

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/util/cli"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemBootstrapperOptions struct {
	XDSAPIImage string
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultEnvoySystemBootstrapper {
	return &DefaultEnvoySystemBootstrapper{
		xdsAPIImage: options.XDSAPIImage,
	}
}

type DefaultEnvoySystemBootstrapper struct {
	xdsAPIImage string
}

func (b *DefaultEnvoySystemBootstrapper) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	namespace := resources.Namespace.Name

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
		},
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
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
			Name:     envoyXDSAPI,
		},
	}

	labels := map[string]string{
		labelKeyEnvoyXDSAPI: "true",
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   envoyXDSAPI,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "envoy-xds-api",
							Args: []string{
								"-v", "5",
								"-logtostderr",
								"-namespace", namespace,
							},
							Image: b.xdsAPIImage,
							Ports: []corev1.ContainerPort{
								{
									HostPort:      8080,
									ContainerPort: 8080,
								},
							},
						},
					},
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: serviceAccount.Name,
					Tolerations: []corev1.Toleration{
						kubeconstants.TolerationNodePool,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &kubeconstants.NodeAffinityNodePool,
					},
				},
			},
		},
	}

	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)
	resources.RoleBindings = append(resources.RoleBindings, roleBinding)
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}

func ParseSystemBootstrapperFlags(vars []string) (*SystemBootstrapperOptions, error) {
	options := &SystemBootstrapperOptions{}
	flags := cli.EmbeddedFlag{
		Target: &options,
		Expected: map[string]cli.EmbeddedFlagValue{
			"xds-api-image": {
				Required:     true,
				EncodingName: "XDSAPIImage",
			},
		},
	}

	err := flags.Parse(vars)
	return options, err
}
