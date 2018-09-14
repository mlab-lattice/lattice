package system

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	"github.com/satori/go.uuid"
)

const (
	serviceScaleRate   = 0.3
	serviceScalePeriod = 2 * time.Second
)

type controller struct {
	backend           *Backend
	actions           *syncutil.LifecycleActionManager
	componentResolver resolver.ComponentResolver
}

func (c *controller) CreateSystem(system *systemRecord) {
	go c.createSystem(system)
}

func (c *controller) RunBuild(build *v1.Build, record *systemRecord) {
	go c.runBuild(build, record)
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

func (c *controller) createSystem(record *systemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("initializing system %v", record.system.ID)

	c.backend.Lock()
	defer c.backend.Unlock()
	record.system.State = v1.SystemStateStable
}

func (c *controller) runBuild(build *v1.Build, record *systemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("evaluating build %v", build.ID)

	if !c.resolveBuildComponent(build, record) {
		return
	}

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		log.Printf("running workload builds for build %v", build.ID)

		now := time.Now()
		build.State = v1.BuildStateRunning
		build.StartTimestamp = &now
		build.Workloads = make(map[tree.Path]v1.WorkloadBuild)

		info := record.builds[build.ID]
		info.Definition.V1().Workloads(func(path tree.Path, workload definitionv1.Workload, info *resolver.ResolutionInfo) tree.WalkContinuation {
			workloadBuild := v1.WorkloadBuild{
				ContainerBuild: v1.ContainerBuild{
					ID:    v1.ContainerBuildID(uuid.NewV4().String()),
					State: v1.ContainerBuildStateRunning,

					StartTimestamp: &now,
				},
				Sidecars: make(map[string]v1.ContainerBuild),
			}

			for name := range workload.Containers().Sidecars {
				workloadBuild.Sidecars[name] = v1.ContainerBuild{
					ID:    v1.ContainerBuildID(uuid.NewV4().String()),
					State: v1.ContainerBuildStateRunning,

					StartTimestamp: &now,
				}
			}

			build.Workloads[path] = workloadBuild
			return tree.ContinueWalk
		})
	}()

	// Wait for builds to complete.
	time.Sleep(10 * time.Second)

	log.Printf("completing build %v", build.ID)

	c.backend.Lock()
	defer c.backend.Unlock()
	now := time.Now()

	// Complete service builds and build.
	for path, workload := range build.Workloads {
		workload.ContainerBuild = v1.ContainerBuild{
			ID:    workload.ID,
			State: v1.ContainerBuildStateSucceeded,

			StartTimestamp:      workload.StartTimestamp,
			CompletionTimestamp: &now,
		}

		for sidecar, build := range workload.Sidecars {
			workload.Sidecars[sidecar] = v1.ContainerBuild{
				ID:    build.ID,
				State: v1.ContainerBuildStateSucceeded,

				StartTimestamp:      build.StartTimestamp,
				CompletionTimestamp: &now,
			}
		}

		build.Workloads[path] = workload
	}

	build.State = v1.BuildStateSucceeded
	build.CompletionTimestamp = &now

	log.Printf("build %v complete", build.ID)
}

func (c *controller) runDeploy(deploy *v1.Deploy, record *systemRecord) {
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
	defer c.actions.ReleaseDeploy(record.system.ID, deploy.ID)

	if deploy.Build == nil {
		if !c.createDeployBuild(deploy, record) {
			return
		}
	}

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		log.Printf("running deploy %v", deploy.ID)

		now := time.Now()
		deploy.State = v1.DeployStateAccepted
		deploy.StartTimestamp = &now
	}()

	if !c.waitForBuildTermination(deploy, record) {
		return
	}

	c.deployBuild(deploy, path, record)

	c.backend.Lock()
	defer c.backend.Unlock()
	now := time.Now()

	deploy.State = v1.DeployStateSucceeded
	deploy.CompletionTimestamp = &now

	log.Printf("deploy %v complete", deploy.ID)
}

func (c *controller) runTeardown(teardown *v1.Teardown, record *systemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("evaluating teardown %v", teardown.ID)

	if !c.lockTeardown(teardown, record) {
		return
	}
	defer c.actions.ReleaseTeardown(record.system.ID, teardown.ID)

	var wg sync.WaitGroup

	// tear down services
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		now := time.Now()
		teardown.StartTimestamp = &now
		teardown.State = v1.TeardownStateInProgress

		record.definition.V1().Services(func(path tree.Path, _ *definitionv1.Service, _ *resolver.ResolutionInfo) tree.WalkContinuation {
			wg.Add(1)
			go c.terminateService(path, record, &wg)
			return tree.ContinueWalk
		})
	}()

	wg.Wait()

	// tear down node pools
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		for subcomponent := range record.nodePools {
			wg.Add(1)
			go c.terminateNodePool(subcomponent, record, &wg)
		}
	}()

	wg.Wait()

	c.backend.Lock()
	defer c.backend.Unlock()

	record.definition = resolver.NewResolutionTree()
	now := time.Now()
	teardown.CompletionTimestamp = &now
	teardown.State = v1.TeardownStateSucceeded
}

func (c *controller) runJob(job *v1.Job) {
	// add a little artificial delay before starting
	time.Sleep(time.Second)

	log.Printf("running job %v", job.ID)

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

func (c *controller) resolveBuildComponent(build *v1.Build, record *systemRecord) bool {
	var buildInfo *buildInfo
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()
		buildInfo = record.builds[build.ID]
	}()

	log.Printf("getting component for build %v", build.ID)

	path, cmpnt, ctx, ok := c.getBuildComponent(buildInfo.Build, record)
	if !ok {
		return false
	}

	log.Printf("resolving definition for build %v", build.ID)

	t, err := c.componentResolver.Resolve(cmpnt, record.system.ID, path, ctx, resolver.DepthInfinite)
	c.backend.Lock()
	defer c.backend.Unlock()

	if err != nil {
		build.State = v1.BuildStateFailed
		build.Message = fmt.Sprintf("error resolving system: %v", err)
		return false
	}

	// ensure that the component is a system if it's at the root
	if path.IsRoot() {
		root, ok := t.Get(tree.RootPath())
		if !ok {
			buildInfo.Build.State = v1.BuildStateFailed
			buildInfo.Build.Message = "system does not have root"
			return false
		}

		_, ok = root.Component.(*definitionv1.System)
		if !ok {
			buildInfo.Build.State = v1.BuildStateFailed
			buildInfo.Build.Message = "root component must be a system"
			return false
		}
	}

	buildInfo.Definition = t
	buildInfo.Build.State = v1.BuildStateAccepted
	return true
}

func (c *controller) getBuildComponent(
	build *v1.Build,
	record *systemRecord,
) (tree.Path, component.Interface, *git.CommitReference, bool) {
	c.backend.Lock()
	defer c.backend.Unlock()

	if build.Path == nil {
		tag := string(*build.Version)
		ref := &definitionv1.Reference{
			GitRepository: &definitionv1.GitRepositoryReference{
				GitRepository: &definitionv1.GitRepository{
					URL: record.system.DefinitionURL,
					Tag: &tag,
				},
			},
		}

		return tree.RootPath(), ref, nil, true
	}

	path := *build.Path
	if record.definition == nil {
		build.State = v1.BuildStateFailed
		build.Message = fmt.Sprintf("system %v does not have any components, cannot build the system based off a path", record.system.ID)
		return "", nil, nil, false
	}

	if path == tree.RootPath() {
		info, ok := record.definition.Get(path)
		if !ok {
			build.State = v1.BuildStateFailed
			build.Message = fmt.Sprintf("system %v does not contain %v", record.system.ID, path.String())
			return "", nil, nil, false
		}

		return path, info.Component, info.Commit, true
	}

	name, _ := path.Leaf()
	parent, _ := path.Parent()
	parentInfo, ok := record.definition.Get(parent)
	if !ok {
		build.State = v1.BuildStateFailed
		build.Message = fmt.Sprintf("system %v does not contain %v", record.system.ID, path.String())
		return "", nil, nil, false
	}

	s, ok := parentInfo.Component.(*definitionv1.System)
	if !ok {
		build.State = v1.BuildStateFailed
		build.Message = fmt.Sprintf("system %v internal node %v is not a system", record.system.ID, parent.String())
		return "", nil, nil, false
	}

	cmpnt, ok := s.Components[name]
	if !ok {
		build.State = v1.BuildStateFailed
		build.Message = fmt.Sprintf("system %v does not contain %v", record.system.ID, path.String())
		return "", nil, nil, false
	}

	return path, cmpnt, parentInfo.Commit, true
}

func (c *controller) getDeployPath(deploy *v1.Deploy, record *systemRecord) (tree.Path, bool) {
	c.backend.Lock()
	defer c.backend.Unlock()

	// get the deploy's path so we can attempt to acquire the proper lifecycle lock
	if deploy.Build != nil {
		buildID := *deploy.Build
		buildInfo, ok := record.builds[buildID]
		if !ok {
			deploy.State = v1.DeployStateFailed
			deploy.Message = fmt.Sprintf("deploy %v build %v does not exist", deploy.ID, buildID)
			return "", false
		}

		// There are many factors that could make an old partial system build incompatible with the
		// current state of the system. Instead of trying to enumerate and handle them, for now
		// we'll simply fail the deploy.
		// May want to revisit this.
		if buildInfo.Build.Path != nil {
			deploy.State = v1.DeployStateFailed
			deploy.Message = fmt.Sprintf("cannot deploy using a build id (%v) since it is only a partial system build", buildID)
			return "", false
		}

		return tree.RootPath(), true
	}

	if deploy.Path != nil {
		return *deploy.Path, true
	}

	return tree.RootPath(), true
}

func (c *controller) lockDeploy(deploy *v1.Deploy, path tree.Path, record *systemRecord) bool {
	c.backend.Lock()
	defer c.backend.Unlock()

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err := c.actions.AcquireDeploy(record.system.ID, deploy.ID, path)
	if err != nil {
		deploy.State = v1.DeployStateFailed
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			deploy.Message = err.Error()
			return false
		}

		deploy.Message = fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error())
		return false
	}

	deploy.State = v1.DeployStateAccepted
	return true
}

func (c *controller) createDeployBuild(deploy *v1.Deploy, record *systemRecord) bool {
	log.Printf("creating build for deploy %v", deploy.ID)

	var build *v1.Build
	var err error
	if deploy.Path != nil {
		build, err = c.backend.Builds(record.system.ID).CreateFromPath(*deploy.Path)
	} else {
		build, err = c.backend.Builds(record.system.ID).CreateFromVersion(*deploy.Version)
	}

	c.backend.Lock()
	defer c.backend.Unlock()

	if err != nil {
		log.Printf("build cretion for deploy %v failed: %v", deploy.ID, err)
		deploy.State = v1.DeployStateFailed
		deploy.Message = fmt.Sprintf("failed to create build: %v", err)
		return false
	}

	log.Printf("creating build %v for deploy %v", build.ID, deploy.ID)
	deploy.Build = &build.ID
	return true
}

func (c *controller) waitForBuildTermination(deploy *v1.Deploy, record *systemRecord) bool {
	log.Printf("waiting for build for deploy %v to terminate", deploy.ID)
	for {
		build, err := c.backend.Builds(record.system.ID).Get(*deploy.Build)
		if err != nil {
			deploy.State = v1.DeployStateFailed
			return false
		}

		switch build.State {
		case v1.BuildStateSucceeded:
			func() {
				c.backend.Lock()
				defer c.backend.Unlock()
				log.Printf("build %v for deploy %v succeeded, moving deploy to in progress", deploy.Build, deploy.ID)
				deploy.State = v1.DeployStateInProgress
			}()
			return true

		case v1.BuildStateFailed:
			func() {
				c.backend.Lock()
				defer c.backend.Unlock()
				log.Printf("build %v for deploy %v failed, failing deploy", deploy.Build, deploy.ID)
				deploy.State = v1.DeployStateFailed
			}()
			return false
		}

		time.Sleep(time.Second)
	}
}

func (c *controller) deployBuild(deploy *v1.Deploy, path tree.Path, record *systemRecord) {
	var definition *resolver.ResolutionTree

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		definition = c.backend.registry[record.system.ID].builds[*deploy.Build].Definition
	}()

	var wg sync.WaitGroup

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		log.Printf("updating existing services")

		// act on existing services
		record.definition.V1().Services(func(path tree.Path, service *definitionv1.Service, info *resolver.ResolutionInfo) tree.WalkContinuation {
			wg.Add(1)

			// if the path is no longer in the tree, terminate the service
			other, ok := definition.Get(path)
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
		definition.V1().Services(func(path tree.Path, service *definitionv1.Service, info *resolver.ResolutionInfo) tree.WalkContinuation {
			// if this path is already a service, then we would have already addressed it above
			if _, ok := record.servicePaths[path]; ok {
				return tree.ContinueWalk
			}

			wg.Add(1)
			go c.addService(path, service, record, &wg)
			return tree.ContinueWalk
		})

		// index the node pools in our build
		nodePools := make(map[tree.PathSubcomponent]*definitionv1.NodePool)
		definition.V1().NodePools(func(subcomponent tree.PathSubcomponent, pool *definitionv1.NodePool) tree.WalkContinuation {
			nodePools[subcomponent] = pool
			return tree.ContinueWalk
		})

		log.Printf("updating existing node pools")

		// handle existing node pools
		for subcomponent, nodePool := range record.nodePools {
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

		log.Printf("replacing system %v definition at %v", record.system.ID, path.String())

		record.definition.ReplacePrefix(path, definition)
	}()

	log.Print("waiting for deploy to finish")
	wg.Wait()
}

func (c *controller) lockTeardown(teardown *v1.Teardown, record *systemRecord) bool {
	c.backend.Lock()
	defer c.backend.Unlock()

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err := c.actions.AcquireTeardown(record.system.ID, teardown.ID)
	if err != nil {
		teardown.State = v1.TeardownStateFailed
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			teardown.Message = err.Error()
			return false
		}

		teardown.Message = fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error())
		return false
	}

	teardown.State = v1.TeardownStateInProgress
	return true
}

func (c *controller) addService(path tree.Path, definition *definitionv1.Service, record *systemRecord, wg *sync.WaitGroup) {
	log.Printf("adding service %v for system %v", path.String(), record.system.ID)

	defer wg.Done()

	var service *v1.Service

	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		service = &v1.Service{
			ID:   v1.ServiceID(uuid.NewV4().String()),
			Path: path,

			State: v1.ServiceStatePending,

			AvailableInstances:   0,
			UpdatedInstances:     0,
			StaleInstances:       0,
			TerminatingInstances: 0,

			Ports: make(map[int32]string),

			Instances: make([]string, 0),
		}

		record.services[service.ID] = &serviceInfo{
			Service:    service,
			Definition: definition,
		}

		record.servicePaths[path] = service.ID
	}()

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.backend.Lock()
			defer c.backend.Unlock()

			log.Printf("scaling service %v for system %v", path.String(), record.system.ID)

			// increase the total number of instances available as a factor of the desired number
			available := int32(math.Min(
				math.Ceil(float64((1+serviceScaleRate)*float64(definition.NumInstances))),
				float64(definition.NumInstances),
			))

			service.AvailableInstances = available
			service.UpdatedInstances = available

			// add new instance ids
			diff := available - service.AvailableInstances
			var newInstances []string
			for i := int32(0); i < diff; i++ {
				newInstances = append(newInstances, uuid.NewV4().String())
			}

			service.Instances = append(service.Instances, newInstances...)

			// if we've reached the desired number of instances, we're done
			return service.AvailableInstances == definition.NumInstances
		}()
		if done {
			c.backend.Lock()
			defer c.backend.Unlock()

			service.State = v1.ServiceStateStable

			log.Printf("done scaling service %v for system %v", path.String(), record.system.ID)
			return
		}
	}
}

func (c *controller) rollService(path tree.Path, definition *definitionv1.Service, record *systemRecord, wg *sync.WaitGroup) {
	log.Printf("beginning rolling scaling service %v for system %v", path.String(), record.system.ID)

	defer wg.Done()

	var service *v1.Service
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		id := record.servicePaths[path]
		service = record.services[id].Service
		service.State = v1.ServiceStateUpdating
		service.StaleInstances = service.AvailableInstances
		service.UpdatedInstances = 0
	}()

	desired := definition.NumInstances
	var newInstances []string
	for i := int32(0); i < desired; i++ {
		newInstances = append(newInstances, uuid.NewV4().String())
	}

	var oldInstances []string
	for _, i := range service.Instances {
		oldInstances = append(oldInstances, i)
	}

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.backend.Lock()
			defer c.backend.Unlock()

			log.Printf("rolling scaling service %v for system %v", path.String(), record.system.ID)

			rev := int32(math.Ceil(float64(definition.NumInstances) * serviceScaleRate))
			service.UpdatedInstances = int32(math.Min(float64(service.UpdatedInstances+rev), float64(desired)))
			if service.StaleInstances == 0 {
				service.TerminatingInstances = int32(math.Max(float64(service.TerminatingInstances-rev), 0))
			} else {
				service.StaleInstances = int32(math.Max(float64(service.TerminatingInstances-rev), 0))
			}

			service.AvailableInstances = service.UpdatedInstances + service.StaleInstances
			service.Instances = append(
				oldInstances[:(service.StaleInstances+service.TerminatingInstances)],
				newInstances[:service.UpdatedInstances]...,
			)

			return service.UpdatedInstances == desired
		}()
		if done {
			c.backend.Lock()
			defer c.backend.Unlock()

			service.State = v1.ServiceStateStable

			log.Printf("done rolling service %v for system %v", path.String(), record.system.ID)
			return
		}
	}
}

func (c *controller) terminateService(path tree.Path, record *systemRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("beginning terminating service %v for system %v", path.String(), record.system.ID)

	var service *v1.Service
	func() {
		c.backend.Lock()
		defer c.backend.Unlock()

		id := record.servicePaths[path]
		service = record.services[id].Service
		service.State = v1.ServiceStateDeleting
	}()

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.backend.Lock()
			defer c.backend.Unlock()

			log.Printf("terminating service %v for system %v", path.String(), record.system.ID)

			removeTerminating := int32(math.Ceil(float64(service.TerminatingInstances) * serviceScaleRate))
			remainingTerminating := service.TerminatingInstances - removeTerminating + service.AvailableInstances + service.StaleInstances
			service.AvailableInstances = 0
			service.UpdatedInstances = 0
			service.StaleInstances = 0
			service.TerminatingInstances = remainingTerminating
			service.Instances = service.Instances[:remainingTerminating]

			return service.TerminatingInstances == 0
		}()
		if done {
			c.backend.Lock()
			defer c.backend.Unlock()

			log.Printf("done terminating service %v for system %v", path.String(), record.system.ID)

			delete(record.services, service.ID)
			delete(record.servicePaths, path)

			return
		}
	}
}

func (c *controller) addNodePool(subcomponent tree.PathSubcomponent, definition *definitionv1.NodePool, record *systemRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("adding node pool %v for system %v", subcomponent.String(), record.system.ID)

	c.backend.Lock()
	defer c.backend.Unlock()

	record.nodePools[subcomponent] = &v1.NodePool{
		InstanceType: definition.InstanceType,
		NumInstances: definition.NumInstances,
	}

	// TODO: add node pool scaling
}

func (c *controller) rollNodePool(subcomponent tree.PathSubcomponent, definition *definitionv1.NodePool, record *systemRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("rolling node pool %v for system %v", subcomponent.String(), record.system.ID)

	c.backend.Lock()
	defer c.backend.Unlock()

	record.nodePools[subcomponent] = &v1.NodePool{
		InstanceType: definition.InstanceType,
		NumInstances: definition.NumInstances,
	}

	// TODO: add node pool scaling
}

func (c *controller) terminateNodePool(subcomponent tree.PathSubcomponent, record *systemRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("terminating node pool %v for system %v", subcomponent.String(), record.system.ID)

	c.backend.Lock()
	defer c.backend.Unlock()

	delete(record.nodePools, subcomponent)

	// TODO: add node pool scaling
}
