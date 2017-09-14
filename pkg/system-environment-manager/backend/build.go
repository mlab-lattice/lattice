package backend

import (
	"fmt"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (kb *KubernetesBackend) BuildSystem(latticeNamespace coretypes.LatticeNamespace, system *systemdefinition.System, version string) (string, error) {
	systemBuild, err := getSystemBuild(latticeNamespace, system, version)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemBuild{}
	err = kb.LatticeResourceRestClient.Post().
		Namespace(constants.InternalNamespace).
		Resource(crv1.SystemBuildResourcePlural).
		Body(systemBuild).
		Do().
		Into(result)
	fmt.Println(err)
	return result.Name, err
}

func getSystemBuild(latticeNamespace coretypes.LatticeNamespace, system *systemdefinition.System, version string) (*crv1.SystemBuild, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel: string(latticeNamespace),
		crv1.SystemVersionLabelKey:      version,
	}

	root, err := systemtree.NewNode(systemdefinition.Interface(system), nil)
	if err != nil {
		return nil, err
	}

	services := map[systemtree.NodePath]crv1.SystemBuildServicesInfo{}
	for path, svcNode := range root.Services() {
		services[path] = crv1.SystemBuildServicesInfo{
			Definition: *(svcNode.Definition().(*systemdefinition.Service)),
		}
	}

	sysBuild := &crv1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   string(uuid.NewUUID()),
			Labels: labels,
		},
		Spec: crv1.SystemBuildSpec{
			LatticeNamespace: latticeNamespace,
			Definition:       *system,
			Services:         services,
		},
		Status: crv1.SystemBuildStatus{
			State: crv1.SystemBuildStatePending,
		},
	}

	return sysBuild, nil
}
