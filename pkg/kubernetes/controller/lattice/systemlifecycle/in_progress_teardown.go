package systemlifecycle

import (
	"fmt"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"github.com/golang/glog"
)

func (slc *SystemLifecycleController) syncInProgressTeardown(syst *crv1.SystemTeardown) error {
	system, err := slc.getSystemForTeardown(syst)
	if err != nil {
		return err
	}

	if system == nil {
		newState := crv1.SystemTeardownStatus{
			State: crv1.SystemTeardownStateSucceeded,
		}
		_, err := slc.updateSystemTeardownStatus(syst, newState)
		if err != nil {
			return err
		}

		return slc.relinquishOwningTeardownClaim(syst)
	}

	if system.DeletionTimestamp != nil {
		glog.V(4).Infof("System %v still deleting, requeueing in 30 seconds", system.Name)
		slc.teardownQueue.AddAfter(syst.Namespace+"/"+syst.Name, 30*time.Second)
		return nil
	}

	return slc.latticeResourceClient.Delete().
		Namespace(string(syst.Spec.LatticeNamespace)).
		Resource(crv1.ResourcePluralSystem).
		Name(string(syst.Spec.LatticeNamespace)).
		Do().
		Into(nil)
}

func (slc *SystemLifecycleController) getSystemForTeardown(syst *crv1.SystemTeardown) (*crv1.System, error) {
	var system *crv1.System

	latticeNamespace := syst.Spec.LatticeNamespace
	for _, sysObj := range slc.systemStore.List() {
		sys := sysObj.(*crv1.System)

		if string(latticeNamespace) == sys.Namespace {
			if system != nil {
				return nil, fmt.Errorf("LatticeNamespace %v contains multiple Systems", latticeNamespace)
			}

			system = sys
		}
	}

	return system, nil
}

func (slc *SystemLifecycleController) relinquishOwningTeardownClaim(syst *crv1.SystemTeardown) error {
	slc.owningLifecycleActionsLock.Lock()
	defer slc.owningLifecycleActionsLock.Unlock()

	owningAction, ok := slc.owningLifecycleActions[syst.Spec.LatticeNamespace]
	if !ok {
		return fmt.Errorf("expected teardown %v to be owning action but there was no owning action", syst.Name)
	}

	if owningAction.teardown == nil {
		return fmt.Errorf("expected teardown %v to be owning action but owning action was rollout %v", syst.Name, owningAction.rollout.Name)
	}

	if owningAction.teardown.Name != syst.Name {
		return fmt.Errorf("expected teardown %v to be owning action but owning action was teardown %v", syst.Name, owningAction.teardown.Name)
	}

	delete(slc.owningLifecycleActions, syst.Spec.LatticeNamespace)
	return nil
}
