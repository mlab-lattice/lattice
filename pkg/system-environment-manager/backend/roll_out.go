package backend

import (
	"fmt"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (kb *KubernetesBackend) RollOutSystemBuild(latticeNamespace coretypes.LatticeNamespace, sysBuildId coretypes.SystemBuildId) (string, error) {
	sysBuild, err := kb.getSystemBuildFromId(latticeNamespace, sysBuildId)
	if err != nil {
		return "", err
	}

	sysRollout, err := getSystemRollout(latticeNamespace, sysBuild)
	if err != nil {
		return "", err
	}

	fmt.Printf("%#v", sysRollout)

	result := &crv1.SystemRollout{}
	err = kb.LatticeResourceRestClient.Post().
		Namespace(constants.InternalNamespace).
		Resource(crv1.SystemRolloutResourcePlural).
		Body(sysRollout).
		Do().
		Into(result)
	return result.Name, err
}

func (kb *KubernetesBackend) getSystemBuildFromId(
	latticeNamespace coretypes.LatticeNamespace,
	sysBuildId coretypes.SystemBuildId,
) (*crv1.SystemBuild, error) {
	result := &crv1.SystemBuild{}
	err := kb.LatticeResourceRestClient.Get().
		Namespace(constants.InternalNamespace).
		Resource(crv1.SystemBuildResourcePlural).
		Name(string(sysBuildId)).
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
			Definition:       sysBuild.Spec.Definition,
			BuildName:        sysBuild.Name,
		},
		Status: crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStatePending,
		},
	}

	return sysRollout, nil
}
