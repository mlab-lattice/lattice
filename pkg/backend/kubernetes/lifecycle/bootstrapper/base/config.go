package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) seedConfig() error {
	fmt.Println("Seeding base lattice config")

	namespace := kubeutil.GetFullNamespace(b.Options.KubeNamespacePrefix, kubeconstants.NamespaceLatticeInternal)

	// Create config
	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ConfigGlobal,
			Namespace: namespace,
		},
		Spec: b.Options.Config,
	}

	_, err := b.LatticeClient.LatticeV1().Configs(namespace).Create(config)
	return err
}
