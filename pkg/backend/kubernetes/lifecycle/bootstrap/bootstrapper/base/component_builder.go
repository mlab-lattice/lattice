package base

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrap/bootstrapper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) componentBuilderResources(resources *bootstrapper.Resources) {
	// FIXME: prefix this cluster role with the cluster id so multiple clusters can have different
	// cluster role definitions
	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconstants.InternalComponentComponentBuilder,
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

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
}
