package systemrollout

import (
	"fmt"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNewSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	root, err := systemtree.NewNode(systemdefinition.Interface(&sysRollout.Spec.Definition), nil)
	if err != nil {
		return nil, err
	}

	services := map[systemtree.NodePath]crv1.SystemServicesInfo{}
	for path, service := range root.Services() {
		svcBuildInfo, ok := sysBuild.Spec.Services[path]
		if !ok {
			return nil, fmt.Errorf("SystemBuild does not have expected Service %v", path)
		}

		services[path] = crv1.SystemServicesInfo{
			Definition: *(service.Definition().(*systemdefinition.Service)),
			BuildName:  *svcBuildInfo.ServiceBuildName,
		}
	}

	sys := &crv1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(sysBuild.Spec.LatticeNamespace),
		},
		Spec: crv1.SystemSpec{
			LatticeNamespace: sysBuild.Spec.LatticeNamespace,
			Services:         services,
		},
		Status: crv1.SystemStatus{
			State: crv1.SystemStateRollingOut,
		},
	}
	return sys, nil
}
