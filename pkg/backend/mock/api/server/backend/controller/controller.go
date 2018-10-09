package controller

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"

	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
)

func New(r *registry.Registry, cr resolver.Interface) *Controller {
	return &Controller{
		registry:          r,
		actions:           syncutil.NewLifecycleActionManager(),
		componentResolver: cr,
	}
}

type Controller struct {
	registry          *registry.Registry
	actions           *syncutil.LifecycleActionManager
	componentResolver resolver.Interface
}

func (c *Controller) CreateSystem(system *registry.SystemRecord) {
	go c.createSystem(system)
}

func (c *Controller) RunBuild(build *v1.Build, record *registry.SystemRecord) {
	go c.runBuild(build, record)
}

func (c *Controller) RunDeploy(deploy *v1.Deploy, record *registry.SystemRecord) {
	go c.runDeploy(deploy, record)
}

func (c *Controller) RunJob(job *v1.Job, record *registry.SystemRecord) {
	go c.runJob(job, record)
}

func (c *Controller) RunTeardown(teardown *v1.Teardown, record *registry.SystemRecord) {
	go c.runTeardown(teardown, record)
}
