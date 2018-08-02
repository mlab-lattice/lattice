package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

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

	successfulContainerBuilds map[string]*latticev1.ContainerBuild
	activeContainerBuilds     map[string]*latticev1.ContainerBuild
	failedContainerBuilds     map[string]*latticev1.ContainerBuild

	servicesNeedNewContainerBuilds map[tree.Path]*definitionv1.Service
	jobsNeedNewContainerBuilds     map[tree.Path]*definitionv1.Job

	// Maps a container build's name to its status
	containerBuildStatuses map[string]latticev1.ContainerBuildStatus

	// Maps a container build's name  to the path of services that are using it
	containerBuildServices map[string][]tree.Path

	// Maps a container build's name  to the path of jobs that are using it
	containerBuildJobs map[string][]tree.Path
}

func (c *Controller) calculateState(build *latticev1.Build) (stateInfo, error) {
	successfulContainerBuilds := make(map[string]*latticev1.ContainerBuild)
	activeContainerBuilds := make(map[string]*latticev1.ContainerBuild)
	failedContainerBuilds := make(map[string]*latticev1.ContainerBuild)
	servicesNeedNewContainerBuilds := make(map[tree.Path]*definitionv1.Service)
	jobsNeedNewContainerBuilds := make(map[tree.Path]*definitionv1.Job)

	containerBuildStatuses := make(map[string]latticev1.ContainerBuildStatus)
	containerBuildServices := make(map[string][]tree.Path)
	containerBuildJobs := make(map[string][]tree.Path)

	// TODO: think about refactoring this and jobs to DRY it up
	for path, service := range build.Spec.Definition.AllServices() {
		serviceInfo, ok := build.Status.Services[path]
		// If the service doesn't have build info yet, note that and continue
		if !ok {
			servicesNeedNewContainerBuilds[path] = service.Service()
			continue
		}

		// Grab all of the container builds for this service
		containerBuildNames := []string{serviceInfo.MainContainer}
		for _, containerBuildName := range serviceInfo.Sidecars {
			containerBuildNames = append(containerBuildNames, containerBuildName)

			services, ok := containerBuildServices[containerBuildName]
			if !ok {
				services = make([]tree.Path, 0)
			}

			services = append(services, path)
			containerBuildServices[containerBuildName] = services
		}

		err := c.updateContainerBuildStatuses(
			build,
			containerBuildNames,
			containerBuildStatuses,
			activeContainerBuilds,
			failedContainerBuilds,
			successfulContainerBuilds,
		)
		if err != nil {
			return stateInfo{}, err
		}
	}

	for path, job := range build.Spec.Definition.AllJobs() {
		jobInfo, ok := build.Status.Jobs[path]
		// If the job doesn't have build info yet, note that and continue
		if !ok {
			jobsNeedNewContainerBuilds[path] = job.Job()
			continue
		}

		// Grab all of the container builds for this service
		containerBuildNames := []string{jobInfo.MainContainer}
		for _, containerBuildName := range jobInfo.Sidecars {
			containerBuildNames = append(containerBuildNames, containerBuildName)

			jobs, ok := containerBuildJobs[containerBuildName]
			if !ok {
				jobs = make([]tree.Path, 0)
			}

			jobs = append(jobs, path)
			containerBuildJobs[containerBuildName] = jobs
		}

		err := c.updateContainerBuildStatuses(
			build,
			containerBuildNames,
			containerBuildStatuses,
			activeContainerBuilds,
			failedContainerBuilds,
			successfulContainerBuilds,
		)
		if err != nil {
			return stateInfo{}, err
		}
	}

	stateInfo := stateInfo{
		successfulContainerBuilds: successfulContainerBuilds,
		activeContainerBuilds:     activeContainerBuilds,
		failedContainerBuilds:     failedContainerBuilds,

		servicesNeedNewContainerBuilds: servicesNeedNewContainerBuilds,
		jobsNeedNewContainerBuilds:     jobsNeedNewContainerBuilds,

		containerBuildStatuses: containerBuildStatuses,
		containerBuildServices: containerBuildServices,
		containerBuildJobs:     containerBuildJobs,
	}

	if len(failedContainerBuilds) > 0 {
		stateInfo.state = stateHasFailedContainerBuilds
		return stateInfo, nil
	}

	if len(servicesNeedNewContainerBuilds) > 0 || len(jobsNeedNewContainerBuilds) > 0 {
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
	names []string,
	statuses map[string]latticev1.ContainerBuildStatus,
	activeContainerBuilds map[string]*latticev1.ContainerBuild,
	failedContainerBuilds map[string]*latticev1.ContainerBuild,
	successfulContainerBuilds map[string]*latticev1.ContainerBuild,
) error {
	// Get the current status of each of the container builds for this component
	for _, name := range names {
		// If we've already processed this container build, no need to do so again
		if _, ok := statuses[name]; ok {
			continue
		}

		containerBuild, err := c.containerBuildLister.ContainerBuilds(build.Namespace).Get(name)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			containerBuild, err = c.latticeClient.LatticeV1().ContainerBuilds(build.Namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"%v has container build %v, but container build does not exist",
						build.Description(c.namespacePrefix),
						name,
					)
					return err
				}

				return err
			}
		}

		statuses[name] = containerBuild.Status

		switch containerBuild.Status.State {
		case latticev1.ContainerBuildStatePending, latticev1.ContainerBuildStateQueued, latticev1.ContainerBuildStateRunning:
			activeContainerBuilds[containerBuild.Name] = containerBuild
		case latticev1.ContainerBuildStateFailed:
			failedContainerBuilds[containerBuild.Name] = containerBuild
		case latticev1.ContainerBuildStateSucceeded:
			successfulContainerBuilds[containerBuild.Name] = containerBuild
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
