package controller

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"
)

func (c *Controller) runDeploy(deploy *v1.Deploy, record *registry.SystemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("evaluating deploy %v", deploy.ID)

	path, ok := c.getDeployPath(deploy, record)
	if !ok {
		return
	}

	if !c.lockDeploy(deploy, path, record) {
		return
	}
	defer c.actions.ReleaseDeploy(record.System.ID, deploy.ID)

	if deploy.Build == nil {
		if !c.createDeployBuild(deploy, record) {
			return
		}
	} else {
		func() {
			c.registry.Lock()
			defer c.registry.Unlock()
			deploy.Status.Build = deploy.Build
		}()
	}

	func() {
		c.registry.Lock()
		defer c.registry.Unlock()
		log.Printf("running deploy %v", deploy.ID)

		now := time.Now()
		deploy.Status.State = v1.DeployStateAccepted
		deploy.Status.StartTimestamp = &now
	}()

	if !c.waitForBuildTermination(deploy, record) {
		return
	}

	c.deployBuild(deploy, path, record)

	c.registry.Lock()
	defer c.registry.Unlock()
	now := time.Now()

	deploy.Status.State = v1.DeployStateSucceeded
	deploy.Status.CompletionTimestamp = &now

	log.Printf("deploy %v complete", deploy.ID)
}

func (c *Controller) getDeployPath(deploy *v1.Deploy, record *registry.SystemRecord) (tree.Path, bool) {
	c.registry.Lock()
	defer c.registry.Unlock()

	// get the deploy's path so we can attempt to acquire the proper lifecycle lock
	if deploy.Build != nil {
		buildID := *deploy.Build
		buildInfo, ok := record.Builds[buildID]
		if !ok {
			deploy.Status.State = v1.DeployStateFailed
			deploy.Status.Message = fmt.Sprintf("deploy %v build %v does not exist", deploy.ID, buildID)
			return "", false
		}

		// There are many factors that could make an old partial system build incompatible with the
		// current state of the system. Instead of trying to enumerate and handle them, for now
		// we'll simply fail the deploy.
		// May want to revisit this.
		if buildInfo.Build.Path != nil {
			deploy.Status.State = v1.DeployStateFailed
			deploy.Status.Message = fmt.Sprintf("cannot deploy using a build id (%v) since it is only a partial system build", buildID)
			return "", false
		}

		return tree.RootPath(), true
	}

	if deploy.Path != nil {
		return *deploy.Path, true
	}

	return tree.RootPath(), true
}

func (c *Controller) lockDeploy(deploy *v1.Deploy, path tree.Path, record *registry.SystemRecord) bool {
	c.registry.Lock()
	defer c.registry.Unlock()

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err := c.actions.AcquireDeploy(record.System.ID, deploy.ID, path)
	if err != nil {
		deploy.Status.State = v1.DeployStateFailed
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			deploy.Status.Message = err.Error()
			return false
		}

		deploy.Status.Message = fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error())
		return false
	}

	deploy.Status.State = v1.DeployStateAccepted
	return true
}

func (c *Controller) createDeployBuild(deploy *v1.Deploy, record *registry.SystemRecord) bool {
	log.Printf("creating build for deploy %v", deploy.ID)

	c.registry.Lock()
	defer c.registry.Unlock()

	build := c.registry.CreateBuild(deploy.Path, deploy.Version, record)
	c.RunBuild(build, record)

	log.Printf("creating build %v for deploy %v", build.ID, deploy.ID)
	deploy.Status.Build = &build.ID
	return true
}

func (c *Controller) waitForBuildTermination(deploy *v1.Deploy, record *registry.SystemRecord) bool {
	log.Printf("waiting for build for deploy %v to terminate", deploy.ID)
	for {
		done, ok := func() (bool, bool) {
			c.registry.Lock()
			defer c.registry.Unlock()

			build := record.Builds[*deploy.Status.Build].Build
			deploy.Status.Path = build.Status.Path
			deploy.Status.Version = build.Status.Version

			switch build.Status.State {
			case v1.BuildStateSucceeded:
				log.Printf("build %v for deploy %v succeeded, moving deploy to in progress", deploy.Status.Build, deploy.ID)
				deploy.Status.State = v1.DeployStateInProgress

				// if this is a root deploy update the system's version
				if deploy.Status.Path.IsRoot() {
					record.System.Status.Version = deploy.Status.Version
				}
				return true, true

			case v1.BuildStateFailed:
				log.Printf("build %v for deploy %v failed, failing deploy", deploy.Status.Build, deploy.ID)
				deploy.Status.State = v1.DeployStateFailed
				return true, false
			}

			return false, true
		}()
		if done {
			return ok
		}

		time.Sleep(time.Second)
	}
}

func (c *Controller) deployBuild(deploy *v1.Deploy, path tree.Path, record *registry.SystemRecord) {
	var buildDefinition *resolver.ResolutionTree

	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		buildDefinition = record.Builds[*deploy.Status.Build].Definition
	}()

	var wg sync.WaitGroup

	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		log.Printf("updating existing services")

		// act on existing services
		record.Definition.V1().Services(func(path tree.Path, service *definitionv1.Service, info *resolver.ResolutionInfo) tree.WalkContinuation {
			wg.Add(1)

			// if the path is no longer in the tree, terminate the service
			other, ok := buildDefinition.Get(path)
			if !ok {
				go c.terminateService(path, record, &wg)
				return tree.ContinueWalk
			}

			// if it's being replaced by another service, roll the service over
			if otherService, ok := other.Component.(*definitionv1.Service); ok {
				go c.rollService(path, otherService, record, &wg)
				return tree.ContinueWalk
			}

			// otherwise this path is no longer a service, so terminate it
			go c.terminateService(path, record, &wg)
			return tree.ContinueWalk
		})

		log.Printf("creating new services")

		// find new services to add
		buildDefinition.V1().Services(func(path tree.Path, service *definitionv1.Service, info *resolver.ResolutionInfo) tree.WalkContinuation {
			// if this path is already a service, then we would have already addressed it above
			if _, ok := record.ServicePaths[path]; ok {
				return tree.ContinueWalk
			}

			wg.Add(1)
			go c.addService(path, service, record, &wg)
			return tree.ContinueWalk
		})

		// index the node pools in our build
		nodePools := make(map[tree.PathSubcomponent]*definitionv1.NodePool)
		buildDefinition.V1().NodePools(func(subcomponent tree.PathSubcomponent, pool *definitionv1.NodePool) tree.WalkContinuation {
			nodePools[subcomponent] = pool
			return tree.ContinueWalk
		})

		log.Printf("updating existing node pools")

		// handle existing node pools
		for subcomponent, nodePool := range record.NodePools {
			// only worry about node pools that are in this build
			if !subcomponent.Path().HasPrefix(path) {
				continue
			}

			np, ok := nodePools[subcomponent]
			if !ok {
				wg.Add(1)
				go c.terminateNodePool(subcomponent, record, &wg)
				continue
			}

			if !reflect.DeepEqual(nodePool, np) {
				wg.Add(1)
				go c.rollNodePool(subcomponent, np, record, &wg)
			}

			// remove it from the node pools to create it
			delete(nodePools, subcomponent)
		}

		log.Printf("creating new node pools")

		// create new node pools
		for subcomponent, nodePool := range nodePools {
			wg.Add(1)
			go c.addNodePool(subcomponent, nodePool, record, &wg)
		}

		log.Printf("replacing system %v definition at %v", record.System.ID, path.String())

		record.Definition.ReplacePrefix(path, buildDefinition)
	}()

	log.Print("waiting for deploy to finish")
	wg.Wait()
}
