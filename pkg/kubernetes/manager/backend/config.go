package backend

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	"github.com/mlab-lattice/system/pkg/types"
)

func (kb *KubernetesBackend) GetSystemUrl(ln types.LatticeNamespace) (string, error) {
	result := &crv1.Config{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ConfigResourcePlural).
		Name(constants.ConfigGlobal).
		Do().
		Into(result)

	if err != nil {
		return "", err
	}

	return result.Spec.SystemConfigs[ln].Url, nil
}

func (kb *KubernetesBackend) getSystemIP() (string, error) {
	result := &crv1.Config{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ConfigResourcePlural).
		Name(constants.ConfigGlobal).
		Do().
		Into(result)

	if err != nil {
		return "", err
	}

	return result.Spec.Provider.Local.IP, nil
}
