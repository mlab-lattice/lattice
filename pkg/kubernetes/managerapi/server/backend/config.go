package backend

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) GetSystemURL(ln types.LatticeNamespace) (string, error) {
	config, err := kb.getConfig()
	if err != nil {
		return "", err
	}

	return config.Spec.SystemConfigs[ln].URL, nil
}

func (kb *KubernetesBackend) getSystemIP() (string, error) {
	config, err := kb.getConfig()
	if err != nil {
		return "", err
	}

	return config.Spec.Provider.Local.IP, nil
}

func (kb *KubernetesBackend) getConfig() (*crv1.Config, error) {
	return kb.LatticeClient.V1().Configs(constants.NamespaceLatticeInternal).Get(constants.ConfigGlobal, metav1.GetOptions{})
}
