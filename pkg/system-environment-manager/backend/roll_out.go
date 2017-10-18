package backend

import (
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (kb *KubernetesBackend) RollOutSystem(ln coretypes.LatticeNamespace, sd *systemdefinition.System, v coretypes.SystemVersion) (coretypes.SystemRolloutId, error) {
	bid, err := kb.BuildSystem(ln, sd, v)
	if err != nil {
		return "", err
	}

	return kb.RollOutSystemBuild(ln, bid)
}

func (kb *KubernetesBackend) RollOutSystemBuild(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildId) (coretypes.SystemRolloutId, error) {
	sysBuild, err := kb.getSystemBuildFromId(ln, bid)
	if err != nil {
		return "", err
	}

	sysRollout, err := getSystemRollout(ln, sysBuild)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemRollout{}
	err = kb.LatticeResourceClient.Post().
		Namespace(constants.InternalNamespace).
		Resource(crv1.SystemRolloutResourcePlural).
		Body(sysRollout).
		Do().
		Into(result)

	return coretypes.SystemRolloutId(result.Name), err
}

func (kb *KubernetesBackend) getSystemBuildFromId(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildId) (*crv1.SystemBuild, error) {
	result := &crv1.SystemBuild{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.InternalNamespace).
		Resource(crv1.SystemBuildResourcePlural).
		Name(string(bid)).
		Do().
		Into(result)

	return result, err
}

func getSystemRollout(latticeNamespace coretypes.LatticeNamespace, sysBuild *crv1.SystemBuild) (*crv1.SystemRollout, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel:   string(latticeNamespace),
		crv1.SystemRolloutVersionLabelKey: sysBuild.Labels[crv1.SystemBuildVersionLabelKey],
		crv1.SystemRolloutBuildIdLabelKey: sysBuild.Name,
	}

	sysRollout := &crv1.SystemRollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:   string(uuid.NewUUID()),
			Labels: labels,
		},
		Spec: crv1.SystemRolloutSpec{
			LatticeNamespace: latticeNamespace,
			BuildName:        sysBuild.Name,
		},
		Status: crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStatePending,
		},
	}

	return sysRollout, nil
}
