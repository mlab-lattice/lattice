package controller

import (
	"log"
	"sync"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func (c *Controller) addNodePool(
	subcomponent tree.PathSubcomponent,
	definition *definitionv1.NodePool,
	record *registry.SystemRecord,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	log.Printf("adding node pool %v for system %v", subcomponent.String(), record.System.ID)

	c.registry.Lock()
	defer c.registry.Unlock()

	record.NodePools[subcomponent] = &v1.NodePool{
		InstanceType: definition.InstanceType,
		NumInstances: definition.NumInstances,
	}

	// TODO: add node pool scaling
}

func (c *Controller) rollNodePool(
	subcomponent tree.PathSubcomponent,
	definition *definitionv1.NodePool,
	record *registry.SystemRecord,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	log.Printf("rolling node pool %v for system %v", subcomponent.String(), record.System.ID)

	c.registry.Lock()
	defer c.registry.Unlock()

	record.NodePools[subcomponent] = &v1.NodePool{
		InstanceType: definition.InstanceType,
		NumInstances: definition.NumInstances,
	}

	// TODO: add node pool scaling
}

func (c *Controller) terminateNodePool(
	subcomponent tree.PathSubcomponent,
	record *registry.SystemRecord,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	log.Printf("terminating node pool %v for system %v", subcomponent.String(), record.System.ID)

	c.registry.Lock()
	defer c.registry.Unlock()

	delete(record.NodePools, subcomponent)

	// TODO: add node pool scaling
}
