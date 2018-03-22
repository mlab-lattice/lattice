package base

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"

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
			Name: kubeconstants.ControlPlaneServiceComponentBuilder,
		},
		Rules: componentBuilderRBACPolicyRules,
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
}

var componentBuilderRBACPolicyRules = []rbacv1.PolicyRule{
	// Read and update lattice component builds
	{
		APIGroups: []string{latticev1.GroupName},
		Resources: []string{latticev1.ResourcePluralComponentBuild},
		Verbs:     readAndUpdateVerbs,
	},
}
