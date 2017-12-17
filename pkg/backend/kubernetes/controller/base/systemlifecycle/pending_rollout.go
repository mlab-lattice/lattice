package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncPendingRollout(rollout *crv1.SystemRollout) error {
	glog.V(5).Infof("syncing pending SystemRollout %v/%v", rollout.Namespace, rollout.Name)
	currentOwningAction := c.attemptToClaimRolloutOwningAction(rollout)
	if currentOwningAction != nil {
		glog.V(5).Infof("SystemRollout %v/%v tried unsuccessfully to claim owning lifecycle action state", rollout.Namespace, rollout.Name)
		_, err := c.updateRolloutStatus(rollout, crv1.SystemRolloutStateFailed, fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()))
		return err
	}

	glog.V(5).Infof("SystemRollout %v/%v successfully claimed owning lifecycle action state", rollout.Namespace, rollout.Name)
	_, err := c.updateRolloutStatus(rollout, crv1.SystemRolloutStateAccepted, "")
	return err
}
