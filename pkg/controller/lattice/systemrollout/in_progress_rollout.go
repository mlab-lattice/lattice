package systemrollout

import (
	"fmt"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func (src *SystemRolloutController) syncInProgressRollout(sysRollout *crv1.SystemRollout) error {
	system, err := src.getSystemForRollout(sysRollout)
	if err != nil {
		return err
	}

	if system == nil {
		sysBuild, err := src.getSystemBuildForRollout(sysRollout)
		if err != nil {
			return err
		}

		if sysBuild == nil {
			return fmt.Errorf("SystemRollout %v is %v without a SystemBuild", sysRollout.Name, crv1.SystemRolloutStateInProgress)
		}

		system, err = src.createSystem(sysRollout, sysBuild)
		if err != nil {
			return err
		}
	}

	return src.syncRolloutWithSystem(sysRollout, system)
}

func (src *SystemRolloutController) getSystemForRollout(sysRollout *crv1.SystemRollout) (*crv1.System, error) {
	var system *crv1.System

	latticeNamespace := sysRollout.Spec.LatticeNamespace
	for _, sysObj := range src.systemStore.List() {
		sys := sysObj.(*crv1.System)

		if latticeNamespace == sys.Spec.LatticeNamespace {
			if system != nil {
				return nil, fmt.Errorf("LatticeNamespace %v contains multiple Systems", latticeNamespace)
			}

			system = sys
		}
	}

	return system, nil
}

func (src *SystemRolloutController) createSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	sys, err := getNewSystem(sysRollout, sysBuild)
	if err != nil {
		return nil, err
	}

	result := &crv1.System{}
	err = src.latticeResourceRestClient.Post().
		Namespace(sysRollout.Namespace).
		Resource(crv1.SystemResourcePlural).
		Body(sys).
		Do().
		Into(result)
	return result, err
}

func (src *SystemRolloutController) syncRolloutWithSystem(sysRollout *crv1.SystemRollout, sys *crv1.System) error {
	var newState crv1.SystemRolloutStatus
	switch sys.Status.State {
	case crv1.SystemStateRollingOut:
		return nil
	case crv1.SystemStateRolloutSucceeded:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateSucceeded,
		}
	case crv1.SystemStateRolloutFailed:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateFailed,
		}
	}

	return src.updateStatus(sysRollout, newState)
}
