package system

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
)

func (c *Controller) syncPendingSystem(system *latticev1.System) error {
	c.configLock.RLock()
	defer c.configLock.RUnlock()
	bootstrappers := []bootstrapper.Interface{
		c.serviceMesh,
		// Cloud provider must come last so that the local cloud provider
		// can strip node selectors/affinities off.
		c.cloudProvider,
	}

	_, err := bootstrap.Bootstrap(
		c.namespacePrefix,
		c.latticeID,
		system.V1ID(),
		system.Spec.DefinitionURL,
		bootstrappers,
		c.kubeClient,
	)
	if err != nil {
		return fmt.Errorf("error bootstrapping %v: %v", system.Description(), err)
	}

	_, err = c.updateSystemStatus(
		system,
		latticev1.SystemStateStable,
		system.Status.Services,
	)
	return err
}
