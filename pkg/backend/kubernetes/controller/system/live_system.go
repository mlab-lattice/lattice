package system

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncLiveSystem(system *latticev1.System) error {
	services, err := c.syncSystemServices(system)
	if err != nil {
		return err
	}

	nodePools, err := c.syncSystemNodePools(system)
	if err != nil {
		return err
	}

	return c.syncSystemStatus(system, services, nodePools)
}
