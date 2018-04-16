package base

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) configResources(resources *bootstrapper.Resources) {
	namespace := kubeutil.InternalNamespace(b.LatticeID)

	config := &latticev1.Config{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Config",
			APIVersion: latticev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ConfigGlobal,
			Namespace: namespace,
		},
		Spec: b.Options.Config,
	}

	resources.Config = config
}
