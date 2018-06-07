package base

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) componentBuilderResources(resources *bootstrapper.Resources) {
	name := kubeutil.ComponentBuilderClusterRoleName(b.NamespacePrefix)

	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: componentBuilderRBACPolicyRules,
	}
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
}

var componentBuilderRBACPolicyRules = []rbacv1.PolicyRule{
	// Read and update lattice component builds
	{
		APIGroups: []string{latticev1.GroupName},
		Resources: []string{latticev1.ResourcePluralContainerBuild},
		Verbs:     ReadAndUpdateVerbs,
	},
}
