package system

import (
	"log"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/satori/go.uuid"
)

type controller struct {
	backend *Backend
}

func (c *controller) CreateSystem(system *systemRecord) {
	go c.createSystem(system)
}

func (c *controller) RunBuild(build *v1.Build) {
	go c.runBuild(build)
}

func (c *controller) RunDeploy(deploy *v1.Deploy, record *systemRecord) {
	go c.runDeploy(deploy, record)
}

func (c *controller) RunJob(job *v1.Job) {
	go c.runJob(job)
}

func (c *controller) RunTeardown(teardown *v1.Teardown, record *systemRecord) {
	go c.runTeardown(teardown, record)
}

func (c *controller) createSystem(system *systemRecord) {
	// Add a bit of delay before the controller starts the build.
	time.Sleep(1 * time.Second)

	c.backend.Lock()
	defer c.backend.Unlock()
	system.system.State = v1.SystemStateStable
}

func (c *controller) runBuild(build *v1.Build) {
	log.Printf("beginning to process build %v", build.ID)
	// Add a bit of delay before the controller starts the build.
	time.Sleep(1 * time.Second)

	// Run service builds.
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		log.Printf("running service builds for build %v", build.ID)

		now := time.Now()
		build.State = v1.BuildStateRunning
		build.StartTimestamp = &now

		for sp, s := range build.Services {
			s.State = v1.ContainerBuildStateRunning
			s.StartTimestamp = &now
			s.ContainerBuild.State = v1.ContainerBuildStateRunning
			s.ContainerBuild.StartTimestamp = &now

			build.Services[sp] = s
		}
	}()

	// Wait for service builds to complete.
	time.Sleep(10 * time.Second)

	c.backend.Lock()
	defer c.backend.Unlock()
	log.Printf("completing build %v", build.ID)
	now := time.Now()

	// Complete service builds and build.
	for sp, s := range build.Services {
		s.State = v1.ContainerBuildStateSucceeded
		s.CompletionTimestamp = &now
		s.ContainerBuild.CompletionTimestamp = &now
		s.ContainerBuild.State = v1.ContainerBuildStateSucceeded
		s.CompletionTimestamp = &now
		build.Services[sp] = s
	}

	build.State = v1.BuildStateSucceeded
	build.CompletionTimestamp = &now
}

func (c *controller) runDeploy(deploy *v1.Deploy, record *systemRecord) {
	log.Printf("beginning to process deploy %v", deploy.ID)
	ok := func() bool {
		c.backend.Lock()
		defer c.backend.Unlock()

		// ensure that there is not other deploy accepted/running
		for _, d := range record.deploys {
			if d.State == v1.DeployStateAccepted || d.State == v1.DeployStateInProgress {
				deploy.State = v1.DeployStateFailed
				log.Printf("found conflicting deploy %v, failing deploy %v", d.ID, deploy.ID)
				return false
			}
		}

		deploy.State = v1.DeployStateAccepted
		return true
	}()
	if !ok {
		return
	}

	// wait until deploy goes in progress or succeeds
	log.Printf("waiting for build %v to complete for deploy %v", deploy.BuildID, deploy.ID)
Loop:
	for {
		build, err := c.backend.Builds(record.system.ID).Get(deploy.BuildID)
		if err != nil {
			deploy.State = v1.DeployStateFailed
			return
		}

		switch build.State {
		case v1.BuildStateSucceeded:
			func() {
				c.backend.Lock()
				defer c.backend.Unlock()
				log.Printf("build %v for deploy %v succeeded, moving deploy to in progress", deploy.BuildID, deploy.ID)
				deploy.State = v1.DeployStateInProgress
			}()
			break Loop

		case v1.BuildStateFailed:
			deploy.State = v1.DeployStateFailed
			log.Printf("build %v for deploy %v failed, failing deploy", deploy.BuildID, build.ID)
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	// sleep for 5 seconds then succeed
	time.Sleep(5 * time.Second)

	build, err := c.backend.Builds(record.system.ID).Get(deploy.BuildID)
	if err != nil {
		deploy.State = v1.DeployStateFailed
		return
	}

	services := make(map[tree.Path]v1.Service)
	for path := range build.Services {
		services[path] = v1.Service{
			ID:                 v1.ServiceID(uuid.NewV4().String()),
			State:              v1.ServiceStateStable,
			Instances:          []string{uuid.NewV4().String()},
			Path:               path,
			AvailableInstances: 1,
		}
	}

	c.backend.Lock()
	defer c.backend.Unlock()
	log.Printf("completing deploy %v", deploy.ID)

	record.system.Services = services
	deploy.State = v1.DeployStateSucceeded
}

func (c *controller) runJob(job *v1.Job) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		now := time.Now()
		job.State = v1.JobStateRunning
		job.StartTimestamp = &now
	}()

	// sleep
	time.Sleep(7 * time.Second)

	c.backend.Lock()
	defer c.backend.Unlock()
	now := time.Now()
	job.State = v1.JobStateSucceeded
	job.CompletionTimestamp = &now
}

func (c *controller) runTeardown(teardown *v1.Teardown, record *systemRecord) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		teardown.State = v1.TeardownStateInProgress

		// tear down services
		for path, s := range record.system.Services {
			s.State = v1.ServiceStateDeleting
			record.system.Services[path] = s
		}
	}()

	// sleep
	time.Sleep(7 * time.Second)

	c.backend.Lock()
	defer c.backend.Unlock()

	record.system.Services = make(map[tree.Path]v1.Service)
	teardown.State = v1.TeardownStateSucceeded
}
