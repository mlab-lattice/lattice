package envoy

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/cli"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LatticeBootstrapperOptions struct {
	PrepareImage      string
	Image             string
	RedirectCIDRBlock string
	XDSAPIImage       string
	XDSAPIPort        int32
}

func NewLatticeBootstrapper(options *LatticeBootstrapperOptions) *DefaultEnvoylatticeBootstrapper {
	return &DefaultEnvoylatticeBootstrapper{
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
	prepareImage      string
	image             string
	redirectCIDRBlock string
	xdsAPIImage       string
	xdsAPIPort        int32
}

func (b *DefaultEnvoylatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	for _, daemonSet := range resources.DaemonSets {
		if daemonSet.Name == kubeconstants.ControlPlaneServiceLatticeControllerManager {
			daemonSet.Spec.Template.Spec.Containers[0].Args = append(
				daemonSet.Spec.Template.Spec.Containers[0].Args,
				"--service-mesh", Envoy,
				"--service-mesh-var", fmt.Sprintf("xds-api-image=%v", b.xdsAPIImage),
			)
		}
	}

	// also need to have the controller manager create envoy SAs for the
	// namespace, so need to give the manager API these privileges
	// so kube doesn't deny creating the SA due to privilege escalation
	for _, clusterRole := range resources.ClusterRoles {
		if clusterRole.Name == kubeconstants.ControlPlaneServiceLatticeControllerManager {
			clusterRole.Rules = append(
				clusterRole.Rules,
				envoyRBACPolicyRules...,
			)
		}
	}

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envoyXDSAPI,
		},
		Rules: envoyRBACPolicyRules,
	}

	resources.Config.Spec.ServiceMesh = latticev1.ConfigServiceMesh{
		Envoy: &latticev1.ConfigServiceMeshEnvoy{
			PrepareImage:      b.prepareImage,
			Image:             b.image,
			RedirectCIDRBlock: b.redirectCIDRBlock,
			XDSAPIImage:       b.xdsAPIImage,
			XDSAPIPort:        b.xdsAPIPort,
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
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
