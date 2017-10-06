package systemrollout

import (
	"fmt"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (src *SystemRolloutController) getNewSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	root, err := systemtree.NewNode(systemdefinition.Interface(&sysBuild.Spec.Definition), nil)
	if err != nil {
		return nil, err
	}

	// Create crv1.SystemServicesInfo for each service in the sysBuild.Spec.Definition
	services := map[systemtree.NodePath]crv1.SystemServicesInfo{}
	for path, service := range root.Services() {
		svcBuildInfo, ok := sysBuild.Spec.Services[path]
		if !ok {
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild does not have expected Service %v", path)
		}

		svcBuild, err := src.getSvcBuild(*svcBuildInfo.BuildName)
		if err != nil {
			return nil, err
		}

		// Create crv1.ComponentBuildArtifacts for each Component in the Service
		cBuildArtifacts := map[string]crv1.ComponentBuildArtifacts{}
		for component, cBuildInfo := range svcBuild.Spec.Components {
			if cBuildInfo.BuildName == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("svcBuild %v Component %v does not have a ComponentBuildName", svcBuild.Name, component)
			}

			cBuildName := *cBuildInfo.BuildName
			cBuildKey := svcBuild.Namespace + "/" + cBuildName
			cBuildObj, exists, err := src.componentBuildStore.GetByKey(cBuildKey)

			if err != nil {
				return nil, err
			}

			if !exists {
				// FIXME: send warn event
				return nil, fmt.Errorf("cBuild %v not in cBuild Store", cBuildKey)
			}

			cBuild := cBuildObj.(*crv1.ComponentBuild)

			if cBuild.Spec.Artifacts == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("cBuild %v does not have Artifacts", cBuildKey)
			}
			cBuildArtifacts[component] = *cBuild.Spec.Artifacts
		}

		services[path] = crv1.SystemServicesInfo{
			Definition:              *(service.Definition().(*systemdefinition.Service)),
			ComponentBuildArtifacts: cBuildArtifacts,
		}
	}

	sys := &crv1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(sysBuild.Spec.LatticeNamespace),
		},
		Spec: crv1.SystemSpec{
			Services: services,
		},
		Status: crv1.SystemStatus{
			State: crv1.SystemStateRollingOut,
		},
	}
	return sys, nil
}

func (src *SystemRolloutController) getSvcBuild(svcBuildName string) (*crv1.ServiceBuild, error) {
	svcBuildKey := constants.InternalNamespace + "/" + svcBuildName
	svcBuildObj, exists, err := src.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("ServiceBuild %v is not in ServiceBuild Store", svcBuildKey)
	}

	svcBuild := svcBuildObj.(*crv1.ServiceBuild)
	return svcBuild, nil
}
