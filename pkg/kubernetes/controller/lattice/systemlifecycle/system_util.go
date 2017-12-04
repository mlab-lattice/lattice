package systemlifecycle

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (slc *Controller) getNewSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	sysSpec, err := slc.getNewSystemSpec(sysRollout, sysBuild)
	if err != nil {
		return nil, err
	}

	sys := &crv1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name:       string(sysRollout.Spec.LatticeNamespace),
			Finalizers: []string{constants.KubeFinalizerSystemController},
		},
		Spec: *sysSpec,
		Status: crv1.SystemStatus{
			State: crv1.SystemStateRollingOut,
		},
	}
	return sys, nil
}

func (slc *Controller) getNewSystemSpec(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.SystemSpec, error) {
	// Create crv1.SystemServicesInfo for each service in the sysBuild.Spec.DefinitionRoot
	services := map[tree.NodePath]crv1.SystemServicesInfo{}
	for path, service := range sysBuild.Spec.DefinitionRoot.Services() {
		svcBuildInfo, ok := sysBuild.Spec.Services[path]
		if !ok {
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild does not have expected Service %v", path)
		}

		svcBuild, err := slc.getSvcBuild(*svcBuildInfo.BuildName)
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
			cBuildObj, exists, err := slc.componentBuildStore.GetByKey(cBuildKey)

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
			Definition:              *(service.Definition().(*definition.Service)),
			ComponentBuildArtifacts: cBuildArtifacts,
		}
	}

	sysSpec := &crv1.SystemSpec{
		Services: services,
	}

	return sysSpec, nil
}

func (slc *Controller) getSvcBuild(svcBuildName string) (*crv1.ServiceBuild, error) {
	svcBuildKey := constants.NamespaceLatticeInternal + "/" + svcBuildName
	svcBuildObj, exists, err := slc.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("ServiceBuild %v is not in ServiceBuild Store", svcBuildKey)
	}

	svcBuild := svcBuildObj.(*crv1.ServiceBuild)
	return svcBuild, nil
}

func (slc *Controller) createSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	sys, err := slc.getNewSystem(sysRollout, sysBuild)
	if err != nil {
		return nil, err
	}

	return slc.latticeClient.V1().Systems(string(sysRollout.Spec.LatticeNamespace)).Create(sys)
}

func (slc *Controller) updateSystemSpec(sys *crv1.System, sysSpec *crv1.SystemSpec) (*crv1.System, error) {
	if reflect.DeepEqual(sys.Spec, sysSpec) {
		return sys, nil
	}

	sys.Spec = *sysSpec

	// FIXME: once CustomResources auto increment generation, remove this (and add observedGeneration)
	// https://github.com/kubernetes/community/pull/913
	sys.Status.State = crv1.SystemStateRollingOut

	return slc.latticeClient.V1().Systems(sys.Namespace).Update(sys)
}
