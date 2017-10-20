package backend

import (
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (kb *KubernetesBackend) BuildSystem(ln coretypes.LatticeNamespace, sd *systemdefinition.System, v coretypes.SystemVersion) (coretypes.SystemBuildId, error) {
	systemBuild, err := getSystemBuild(ln, sd, v)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemBuild{}
	err = kb.LatticeResourceClient.Post().
		Namespace(constants.NamespaceInternal).
		Resource(crv1.SystemBuildResourcePlural).
		Body(systemBuild).
		Do().
		Into(result)

	return coretypes.SystemBuildId(result.Name), err
}

func getSystemBuild(ln coretypes.LatticeNamespace, sd *systemdefinition.System, v coretypes.SystemVersion) (*crv1.SystemBuild, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel: string(ln),
		crv1.SystemVersionLabelKey:      string(v),
	}

	root, err := systemtree.NewNode(systemdefinition.Interface(sd), nil)
	if err != nil {
		return nil, err
	}

	services := map[systemtree.NodePath]crv1.SystemBuildServicesInfo{}
	for path, svcNode := range root.Services() {
		services[path] = crv1.SystemBuildServicesInfo{
			Definition: *(svcNode.Definition().(*systemdefinition.Service)),
		}
	}

	sysB := &crv1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   string(uuid.NewUUID()),
			Labels: labels,
		},
		Spec: crv1.SystemBuildSpec{
			LatticeNamespace: ln,
			Definition:       *sd,
			Services:         services,
		},
		Status: crv1.SystemBuildStatus{
			State: crv1.SystemBuildStatePending,
		},
	}

	return sysB, nil
}
