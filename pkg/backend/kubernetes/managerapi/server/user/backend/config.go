package backend

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) getConfig() (*crv1.Config, error) {
	return kb.LatticeClient.LatticeV1().Configs(constants.NamespaceLatticeInternal).Get(constants.ConfigGlobal, metav1.GetOptions{})
}
