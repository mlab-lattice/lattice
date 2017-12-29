package base

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) configResources(resources *bootstrapper.ClusterResources) {
	namespace := kubeutil.InternalNamespace(b.ClusterID)

	config := &crv1.Config{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Config",
			APIVersion: crv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ConfigGlobal,
			Namespace: namespace,
		},
		Spec: b.Options.Config,
	}

	resources.Config = config
}
