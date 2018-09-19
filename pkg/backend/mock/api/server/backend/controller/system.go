package controller

import (
	"log"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
)

func (c *Controller) createSystem(record *registry.SystemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("initializing system %v", record.System.ID)

	c.registry.Lock()
	defer c.registry.Unlock()
	record.System.Status.State = v1.SystemStateStable
}
