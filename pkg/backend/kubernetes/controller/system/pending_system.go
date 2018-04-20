package system

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap"
)

func (c *Controller) syncPendingSystem(system *latticev1.System) error {
	_, err := bootstrap.Bootstrap(
		c.namespacePrefix,
		c.latticeID,
		system.V1ID(),
		system.Spec.DefinitionURL,
		c.systemBootstrappers,
		c.kubeClient,
	)
	if err != nil {
		return err
	}

	_, err = c.updateSystemStatus(
		system,
		latticev1.SystemStateStable,
		system.Status.Services,
	)
	return err
}
