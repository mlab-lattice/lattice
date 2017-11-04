package backend

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"
)

func (kb *KubernetesBackend) GetSystemUrl(ln coretypes.LatticeNamespace) (string, error) {
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

func (kb *KubernetesBackend) getSystemIP() (*string, error) {
	result := &crv1.Config{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ConfigResourcePlural).
		Name(constants.ConfigGlobal).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	return result.Spec.SystemIP, nil
}
