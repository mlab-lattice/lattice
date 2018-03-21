package system

import (
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap"
	"github.com/mlab-lattice/system/pkg/types"
)

func (c *Controller) syncPendingSystem(system *latticev1.System) error {
	_, err := bootstrap.Bootstrap(
		c.latticeID,
		types.SystemID(system.Name),
		system.Spec.DefinitionURL,
		c.systemBootstrappers,
		c.kubeClient,
		c.latticeClient,
	)
	if err != nil {
		return err
	}

	_, err = c.updateSystemStatus(
		system,
		latticev1.SystemStateStable,
		system.Status.Services,
		system.Status.ServiceStatuses,
	)
	return err
}
