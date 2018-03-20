package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncPendingDeploy(deploy *latticev1.Deploy) error {
	glog.V(5).Infof("syncing pending Deploy %v/%v", deploy.Namespace, deploy.Name)
	currentOwningAction := c.attemptToClaimDeployOwningAction(deploy)
	if currentOwningAction != nil {
		glog.V(5).Infof("Deploy %v/%v tried unsuccessfully to claim owning lifecycle action state", deploy.Namespace, deploy.Name)
		_, err := c.updateDeployStatus(deploy, latticev1.DeployStateFailed, fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()))
		return err
	}

	glog.V(5).Infof("Deploy %v/%v successfully claimed owning lifecycle action state", deploy.Namespace, deploy.Name)
	_, err := c.updateDeployStatus(deploy, latticev1.DeployStateAccepted, "")
	return err
}
