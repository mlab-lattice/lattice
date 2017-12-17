package backend

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) GetSystemDefinitionURL(ln types.LatticeNamespace) (string, error) {
	system, err := kb.LatticeClient.LatticeV1().Systems(string(ln)).Get(string(ln), metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return system.Spec.DefinitionURL, nil
}

func (kb *KubernetesBackend) getSystemIP() (string, error) {
	config, err := kb.getConfig()
	if err != nil {
		return "", err
	}

	return config.Spec.Provider.Local.IP, nil
}

func (kb *KubernetesBackend) getConfig() (*crv1.Config, error) {
	return kb.LatticeClient.LatticeV1().Configs(constants.NamespaceLatticeInternal).Get(constants.ConfigGlobal, metav1.GetOptions{})
}
