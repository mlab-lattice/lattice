package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/satori/go.uuid"
)

func (c *Controller) runBuild(build *v1.Build, record *registry.SystemRecord) {
	// add a little artificial delay before starting
	time.Sleep(1 * time.Second)

	log.Printf("evaluating build %v", build.ID)

	if !c.resolveBuildComponent(build, record) {
		return
	}

	func() {
		c.registry.Lock()
		defer c.registry.Unlock()
		log.Printf("running workload builds for build %v", build.ID)

		now := time.Now()
		build.Status.State = v1.BuildStateRunning
		build.Status.StartTimestamp = &now
		build.Status.Workloads = make(map[tree.Path]v1.WorkloadBuild)

		info := record.Builds[build.ID]
		info.Definition.V1().Workloads(func(path tree.Path, workload definitionv1.Workload, info *resolver.ResolutionInfo) tree.WalkContinuation {
			workloadBuild := v1.WorkloadBuild{
				ContainerBuild: v1.ContainerBuild{
					ID: v1.ContainerBuildID(uuid.NewV4().String()),

					Status: v1.ContainerBuildStatus{
						State: v1.ContainerBuildStateRunning,

						StartTimestamp: &now,
					},
				},
				Sidecars: make(map[string]v1.ContainerBuild),
			}

			for name := range workload.Containers().Sidecars {
				workloadBuild.Sidecars[name] = v1.ContainerBuild{
					ID: v1.ContainerBuildID(uuid.NewV4().String()),

					Status: v1.ContainerBuildStatus{
						State: v1.ContainerBuildStateRunning,

						StartTimestamp: &now,
					},
				}
			}

			build.Status.Workloads[path] = workloadBuild
			return tree.ContinueWalk
		})
	}()

	// Wait for builds to complete.
	time.Sleep(10 * time.Second)

	log.Printf("completing build %v", build.ID)

	c.registry.Lock()
	defer c.registry.Unlock()
	now := time.Now()

	// Complete service builds and build.
	for path, workload := range build.Status.Workloads {
		workload.ContainerBuild = v1.ContainerBuild{
			ID: workload.ID,

			Status: v1.ContainerBuildStatus{
				State: v1.ContainerBuildStateSucceeded,

				StartTimestamp:      workload.Status.StartTimestamp,
				CompletionTimestamp: &now,
			},
		}

		for sidecar, build := range workload.Sidecars {
			workload.Sidecars[sidecar] = v1.ContainerBuild{
				ID: build.ID,

				Status: v1.ContainerBuildStatus{
					State: v1.ContainerBuildStateSucceeded,

					StartTimestamp:      workload.Sidecars[sidecar].Status.StartTimestamp,
					CompletionTimestamp: &now,
				},
			}
		}

		build.Status.Workloads[path] = workload
	}

	build.Status.State = v1.BuildStateSucceeded
	build.Status.CompletionTimestamp = &now

	log.Printf("build %v complete", build.ID)
}

func (c *Controller) resolveBuildComponent(build *v1.Build, record *registry.SystemRecord) bool {
	var buildInfo *registry.BuildInfo
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()
		buildInfo = record.Builds[build.ID]
	}()

	log.Printf("getting component for build %v", build.ID)

	path, cmpnt, ctx, ok := c.getBuildComponent(buildInfo.Build, record)
	if !ok {
		return false
	}

	log.Printf("resolving definition for build %v", build.ID)

	t, err := c.componentResolver.Resolve(cmpnt, record.System.ID, path, ctx, resolver.DepthInfinite)
	c.registry.Lock()
	defer c.registry.Unlock()

	if err != nil {
		build.Status.State = v1.BuildStateFailed
		build.Status.Message = fmt.Sprintf("error resolving system: %v", err)
		return false
	}

	// ensure that the component is a system if it's at the root
	if path.IsRoot() {
		root, ok := t.Get(tree.RootPath())
		if !ok {
			buildInfo.Build.Status.State = v1.BuildStateFailed
			buildInfo.Build.Status.Message = "system does not have root"
			return false
		}

		_, ok = root.Component.(*definitionv1.System)
		if !ok {
			buildInfo.Build.Status.State = v1.BuildStateFailed
			buildInfo.Build.Status.Message = "root component must be a system"
			return false
		}
	}

	buildInfo.Definition = t
	buildInfo.Build.Status.State = v1.BuildStateAccepted
	return true
}

// FIXME(kevinrosendahl): most of this is very similar to the k8s build controller, figure out how much can be unified
func (c *Controller) getBuildComponent(
	build *v1.Build,
	record *registry.SystemRecord,
) (
	tree.Path,
	component.Interface,
	*git.CommitReference,
	bool,
) {
	c.registry.Lock()
	defer c.registry.Unlock()

	if build.Path == nil {
		root := tree.RootPath()
		build.Status.Path = &root
		build.Status.Version = build.Version

		tag := string(*build.Version)
		ref := &definitionv1.Reference{
			GitRepository: &definitionv1.GitRepositoryReference{
				GitRepository: &definitionv1.GitRepository{
					URL: record.System.DefinitionURL,
					Tag: &tag,
				},
			},
		}

		return tree.RootPath(), ref, nil, true
	}

	path := *build.Path
	build.Status.Path = build.Path

	if record.Definition == nil {
		build.Status.State = v1.BuildStateFailed
		build.Status.Message = fmt.Sprintf("system %v does not have any components, cannot build the system based off a path", record.System.ID)
		return "", nil, nil, false
	}

	build.Status.Version = record.System.Status.Version

	if path == tree.RootPath() {
		info, ok := record.Definition.Get(path)
		if !ok {
			build.Status.State = v1.BuildStateFailed
			build.Status.Message = fmt.Sprintf("system %v does not contain %v", record.System.ID, path.String())
			return "", nil, nil, false
		}

		return path, info.Component, info.Commit, true
	}

	name, _ := path.Leaf()
	parent, _ := path.Parent()
	parentInfo, ok := record.Definition.Get(parent)
	if !ok {
		build.Status.State = v1.BuildStateFailed
		build.Status.Message = fmt.Sprintf("system %v does not contain %v", record.System.ID, path.String())
		return "", nil, nil, false
	}

	s, ok := parentInfo.Component.(*definitionv1.System)
	if !ok {
		build.Status.State = v1.BuildStateFailed
		build.Status.Message = fmt.Sprintf("system %v internal node %v is not a system", record.System.ID, parent.String())
		return "", nil, nil, false
	}

	cmpnt, ok := s.Components[name]
	if !ok {
		build.Status.State = v1.BuildStateFailed
		build.Status.Message = fmt.Sprintf("system %v does not contain %v", record.System.ID, path.String())
		return "", nil, nil, false
	}

	return path, cmpnt, parentInfo.Commit, true
}
