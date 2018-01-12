package envoy

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterBootstrapperOptions struct {
	PrepareImage      string
	Image             string
	RedirectCIDRBlock string
	XDSAPIImage       string
	XDSAPIPort        int32
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) *DefaultEnvoyClusterBootstrapper {
	return &DefaultEnvoyClusterBootstrapper{
		prepareImage:      options.PrepareImage,
		image:             options.Image,
		redirectCIDRBlock: options.RedirectCIDRBlock,
		xdsAPIImage:       options.XDSAPIImage,
		xdsAPIPort:        options.XDSAPIPort,
	}
}

type DefaultEnvoyClusterBootstrapper struct {
	prepareImage      string
	image             string
	redirectCIDRBlock string
	xdsAPIImage       string
	xdsAPIPort        int32
}

func (b *DefaultEnvoyClusterBootstrapper) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envoyXDSAPI,
		},
		Rules: []rbacv1.PolicyRule{
			// Read kube endpoints
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"endpoints"},
				Verbs:     []string{"get", "watch", "list"},
			},
			// Read lattice services
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralService},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}

	resources.Config.Spec.ServiceMesh = crv1.ConfigServiceMesh{
		Envoy: &crv1.ConfigServiceMeshEnvoy{
			PrepareImage:      b.prepareImage,
			Image:             b.image,
			RedirectCIDRBlock: b.redirectCIDRBlock,
			XDSAPIImage:       b.xdsAPIImage,
			XDSAPIPort:        b.xdsAPIPort,
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
}
