package controller

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"
	timeutil "github.com/mlab-lattice/lattice/pkg/util/time"
)

func (c *Controller) runTeardown(teardown *v1.Teardown, record *registry.SystemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("evaluating teardown %v", teardown.ID)

	if !c.lockTeardown(teardown, record) {
		return
	}
	defer c.actions.ReleaseTeardown(record.System.ID, teardown.ID)

	var wg sync.WaitGroup

	// tear down services
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		teardown.Status.StartTimestamp = timeutil.New(time.Now())
		teardown.Status.State = v1.TeardownStateInProgress

		record.Definition.V1().Services(func(path tree.Path, _ *definitionv1.Service, _ *resolver.ResolutionInfo) tree.WalkContinuation {
			wg.Add(1)
			go c.terminateService(path, record, &wg)
			return tree.ContinueWalk
		})
	}()

	wg.Wait()

	// tear down node pools
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		for subcomponent := range record.NodePools {
			wg.Add(1)
			go c.terminateNodePool(subcomponent, record, &wg)
		}
	}()

	wg.Wait()

	c.registry.Lock()
	defer c.registry.Unlock()

	record.Definition = resolver.NewResolutionTree()
	teardown.Status.CompletionTimestamp = timeutil.New(time.Now())
	teardown.Status.State = v1.TeardownStateSucceeded
}

func (c *Controller) lockTeardown(teardown *v1.Teardown, record *registry.SystemRecord) bool {
	c.registry.Lock()
	defer c.registry.Unlock()

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err := c.actions.AcquireTeardown(record.System.ID, teardown.ID)
	if err != nil {
		teardown.Status.State = v1.TeardownStateFailed
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			teardown.Status.Message = err.Error()
			return false
		}

		teardown.Status.Message = fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error())
		return false
	}

	teardown.Status.State = v1.TeardownStateInProgress
	return true
}
