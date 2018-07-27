package envoy

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LatticeBootstrapperOptions struct {
	PrepareImage      string
	Image             string
	RedirectCIDRBlock string
	XDSAPIVersion     string
	XDSAPIImage       string
	XDSAPIPort        int32
}

func NewLatticeBootstrapper(namespacePrefix string, options *LatticeBootstrapperOptions) *DefaultEnvoylatticeBootstrapper {
	return &DefaultEnvoylatticeBootstrapper{
		namespacePrefix: namespacePrefix,

		prepareImage:      options.PrepareImage,
		image:             options.Image,
		redirectCIDRBlock: options.RedirectCIDRBlock,
		xdsAPIImage:       options.XDSAPIImage,
		xdsAPIPort:        options.XDSAPIPort,
	}
}

func LatticeBootstrapperFlags() (cli.Flags, *LatticeBootstrapperOptions) {
	options := &LatticeBootstrapperOptions{}
	flags := cli.Flags{
		&cli.StringFlag{
			Name:     "prepare-image",
			Required: true,
			Target:   &options.PrepareImage,
		},
		&cli.StringFlag{
			Name:    "image",
			Default: "envoyproxy/envoy-alpine",
			Target:  &options.Image,
		},
		&cli.StringFlag{
			Name:     "redirect-cidr-block",
			Required: true,
			Target:   &options.RedirectCIDRBlock,
		},
		&cli.StringFlag{
			Name:     "xds-api-image",
			Required: true,
			Target:   &options.XDSAPIImage,
		},
		&cli.Int32Flag{
			Name:    "xds-api-port",
			Default: 8080,
			Target:  &options.XDSAPIPort,
		},
	}
	return flags, options
}

type DefaultEnvoylatticeBootstrapper struct {
	namespacePrefix string

	prepareImage      string
	image             string
	redirectCIDRBlock string
	xdsAPIImage       string
	xdsAPIPort        int32
}

func (b *DefaultEnvoylatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	internalNamespace := kubeutil.InternalNamespace(b.namespacePrefix)
	xdsAPIName := fmt.Sprintf("service-mesh-envoy-%v", xdsAPI)

	for _, daemonSet := range resources.DaemonSets {
		if daemonSet.Name == kubeconstants.ControlPlaneServiceLatticeControllerManager {
			daemonSet.Spec.Template.Spec.Containers[0].Args = append(
				daemonSet.Spec.Template.Spec.Containers[0].Args,
				"--service-mesh", Envoy,
				"--service-mesh-var", fmt.Sprintf("redirect-cidr-block=%v", b.redirectCIDRBlock),
				"--service-mesh-var", fmt.Sprintf("xds-api-port=%v", b.xdsAPIPort),
			)
		}
	}

	xdsAPIclusterRoleName := fmt.Sprintf("%v-%v", b.namespacePrefix, xdsAPIName)
	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: xdsAPIclusterRoleName,
		},
		Rules: envoyRBACPolicyRules,
	}
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      xdsAPIName,
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
			Name:      xdsAPIName,
			Namespace: internalNamespace,
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

	labels := map[string]string{
		labelKeyEnvoyXDSAPI: "true",
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      xdsAPIName,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   xdsAPIName,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "envoy-xds-api",
							Args: []string{
								"-v", "5",
								"-logtostderr",
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
						latticev1.AllNodePoolTolleration,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &latticev1.AllNodePoolAffinity,
					},
				},
			},
		},
	}
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)

	resources.Config.Spec.ServiceMesh = latticev1.ConfigServiceMesh{
		Envoy: &latticev1.ConfigServiceMeshEnvoy{
			PrepareImage: b.prepareImage,
			Image:        b.image,
			XDSAPIImage:  b.xdsAPIImage,
		},
	}
}

var envoyRBACPolicyRules = []rbacv1.PolicyRule{
	// Read kube endpoints
	{
		APIGroups: []string{corev1.GroupName},
		Resources: []string{"endpoints"},
		Verbs:     []string{"get", "watch", "list"},
	},
	// Read lattice services
	{
		APIGroups: []string{latticev1.GroupName},
		Resources: []string{latticev1.ResourcePluralService},
		Verbs:     []string{"get", "watch", "list"},
	},
}
