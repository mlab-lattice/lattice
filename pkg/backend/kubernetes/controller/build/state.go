package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type state string

const (
	stateHasFailedContainerBuilds                 state = "has-failed-container-builds"
	stateHasOnlyRunningOrSucceededContainerBuilds state = "has-only-succeeded-or-running-container-builds"
	stateNoFailuresNeedsNewContainerBuilds        state = "no-failures-needs-new-container-builds"
	stateAllContainerBuildsSucceeded              state = "all-container-builds-succeeded"
)

type stateInfo struct {
	state state

	successfulContainerBuilds map[v1.ContainerBuildID]*latticev1.ContainerBuild
	activeContainerBuilds     map[v1.ContainerBuildID]*latticev1.ContainerBuild
	failedContainerBuilds     map[v1.ContainerBuildID]*latticev1.ContainerBuild

	workloadsNeedNewContainerBuilds map[tree.Path]definitionv1.Workload

	// Maps a container build's ID to its status
	containerBuildStatuses map[v1.ContainerBuildID]latticev1.ContainerBuildStatus

	// Maps a container build's name to the path of workloads that are using it
	containerBuildWorkloads map[v1.ContainerBuildID][]tree.Path
}

func (c *Controller) calculateState(build *latticev1.Build) (stateInfo, error) {
	if build.Status.Definition == nil {
		return stateInfo{}, fmt.Errorf("cannot calculate state for build with no definition")
	}

	successfulContainerBuilds := make(map[v1.ContainerBuildID]*latticev1.ContainerBuild)
	activeContainerBuilds := make(map[v1.ContainerBuildID]*latticev1.ContainerBuild)
	failedContainerBuilds := make(map[v1.ContainerBuildID]*latticev1.ContainerBuild)
	workloadsNeedNewContainerBuilds := make(map[tree.Path]definitionv1.Workload)

	containerBuildStatuses := make(map[v1.ContainerBuildID]latticev1.ContainerBuildStatus)
	containerBuildWorkloads := make(map[v1.ContainerBuildID][]tree.Path)

	var err error
	build.Status.Definition.V1().Workloads(func(path tree.Path, workload definitionv1.Workload, info *resolver.ResolutionInfo) bool {
		buildInfo, ok := build.Status.Workloads[path]
		if !ok {
			workloadsNeedNewContainerBuilds[path] = workload
			return true
		}

		workloads, ok := containerBuildWorkloads[buildInfo.MainContainer]
		if !ok {
			workloads = make([]tree.Path, 0)
		}
		workloads = append(workloads, path)
		containerBuildWorkloads[buildInfo.MainContainer] = workloads

		containerBuilds := []v1.ContainerBuildID{buildInfo.MainContainer}
		for _, sidecarBuild := range buildInfo.Sidecars {
			containerBuilds = append(containerBuilds, sidecarBuild)

			workloads, ok := containerBuildWorkloads[sidecarBuild]
			if !ok {
				workloads = make([]tree.Path, 0)
			}
			workloads = append(workloads, path)
			containerBuildWorkloads[buildInfo.MainContainer] = workloads
		}

		err = c.updateContainerBuildStatuses(
			build,
			containerBuilds,
			containerBuildStatuses,
			activeContainerBuilds,
			failedContainerBuilds,
			successfulContainerBuilds,
		)
		if err != nil {
			return false
		}

		return true
	})

	stateInfo := stateInfo{
		successfulContainerBuilds: successfulContainerBuilds,
		activeContainerBuilds:     activeContainerBuilds,
		failedContainerBuilds:     failedContainerBuilds,

		workloadsNeedNewContainerBuilds: workloadsNeedNewContainerBuilds,

		containerBuildStatuses:  containerBuildStatuses,
		containerBuildWorkloads: containerBuildWorkloads,
	}

	if len(failedContainerBuilds) > 0 {
		stateInfo.state = stateHasFailedContainerBuilds
		return stateInfo, nil
	}

	if len(workloadsNeedNewContainerBuilds) > 0 {
		stateInfo.state = stateNoFailuresNeedsNewContainerBuilds
		return stateInfo, nil
	}

	if len(activeContainerBuilds) > 0 {
		stateInfo.state = stateHasOnlyRunningOrSucceededContainerBuilds
		return stateInfo, nil
	}

	stateInfo.state = stateAllContainerBuildsSucceeded
	return stateInfo, nil
}

func (c *Controller) updateContainerBuildStatuses(
	build *latticev1.Build,
	ids []v1.ContainerBuildID,
	statuses map[v1.ContainerBuildID]latticev1.ContainerBuildStatus,
	activeContainerBuilds map[v1.ContainerBuildID]*latticev1.ContainerBuild,
	failedContainerBuilds map[v1.ContainerBuildID]*latticev1.ContainerBuild,
	successfulContainerBuilds map[v1.ContainerBuildID]*latticev1.ContainerBuild,
) error {
	// Get the current status of each of the container builds for this component
	for _, id := range ids {
		// If we've already processed this container build, no need to do so again
		if _, ok := statuses[id]; ok {
			continue
		}

		containerBuild, err := c.containerBuildLister.ContainerBuilds(build.Namespace).Get(string(id))
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			containerBuild, err = c.latticeClient.LatticeV1().ContainerBuilds(build.Namespace).Get(string(id), metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"%v has container build %v, but container build does not exist",
						build.Description(c.namespacePrefix),
						id,
					)
					return err
				}

				return err
			}
		}

		statuses[id] = containerBuild.Status

		switch containerBuild.Status.State {
		case latticev1.ContainerBuildStatePending, latticev1.ContainerBuildStateQueued, latticev1.ContainerBuildStateRunning:
			activeContainerBuilds[v1.ContainerBuildID(containerBuild.Name)] = containerBuild
		case latticev1.ContainerBuildStateFailed:
			failedContainerBuilds[v1.ContainerBuildID(containerBuild.Name)] = containerBuild
		case latticev1.ContainerBuildStateSucceeded:
			successfulContainerBuilds[v1.ContainerBuildID(containerBuild.Name)] = containerBuild
		default:
			// FIXME: send warn event
			err := fmt.Errorf(
				"%v has unexpected state %v",
				containerBuild.Description(c.namespacePrefix),
				containerBuild.Status.State,
			)
			return err
		}
	}

	return nil
}
