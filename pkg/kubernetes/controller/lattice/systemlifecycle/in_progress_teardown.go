package systemlifecycle

import (
	"fmt"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (slc *Controller) syncInProgressTeardown(syst *crv1.SystemTeardown) error {
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

	ns := string(syst.Spec.LatticeNamespace)
	return slc.latticeClient.V1().Systems(ns).Delete(ns, &metav1.DeleteOptions{})
}

func (slc *Controller) getSystemForTeardown(syst *crv1.SystemTeardown) (*crv1.System, error) {
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

func (slc *Controller) relinquishOwningTeardownClaim(syst *crv1.SystemTeardown) error {
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
