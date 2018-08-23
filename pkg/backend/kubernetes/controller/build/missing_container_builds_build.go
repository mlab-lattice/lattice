package build

import (
	"encoding/json"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/sha1"

	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncMissingContainerBuildsBuild(build *latticev1.Build, stateInfo stateInfo) error {
	containerBuildStatuses := stateInfo.containerBuildStatuses
	containerBuildHashes := make(map[string]*latticev1.ContainerBuild)
	services := make(map[tree.Path]latticev1.BuildStatusService)
	jobs := make(map[tree.Path]latticev1.BuildStatusJob)

	// look through all the containers of each service to see if there are any containers that
	// don't have builds yet
	// TODO: think about how to refactor this to DRY it up
	for path, service := range stateInfo.servicesNeedNewContainerBuilds {
		containers := map[string]definitionv1.Container{
			kubeutil.UserMainContainerName: service.Container,
		}
		for name, sidecarContainer := range service.Sidecars {
			containers[kubeutil.UserSidecarContainerName(name)] = sidecarContainer
		}

		// maps the service's container names to the container builds for them
		containerBuilds := make(map[string]string)
		err := c.getContainerBuilds(build, path, containers, containerBuilds, containerBuildHashes, containerBuildStatuses)
		if err != nil {
			return nil
		}

		statusServiceInfo := latticev1.BuildStatusService{
			MainContainer: containerBuilds[kubeutil.UserMainContainerName],
			Sidecars:      make(map[string]string),
		}
		for sidecar := range service.Sidecars {
			statusServiceInfo.Sidecars[sidecar] = containerBuilds[kubeutil.UserSidecarContainerName(sidecar)]
		}
		services[path] = statusServiceInfo
	}

	for path, job := range stateInfo.jobsNeedNewContainerBuilds {
		containers := map[string]definitionv1.Container{
			kubeutil.UserMainContainerName: job.Container,
		}
		for name, sidecarContainer := range job.Sidecars {
			containers[kubeutil.UserSidecarContainerName(name)] = sidecarContainer
		}

		// maps the service's container names to the container builds for them
		containerBuilds := make(map[string]string)
		err := c.getContainerBuilds(build, path, containers, containerBuilds, containerBuildHashes, containerBuildStatuses)
		if err != nil {
			return nil
		}

		statusJobInfo := latticev1.BuildStatusJob{
			MainContainer: containerBuilds[kubeutil.UserMainContainerName],
			Sidecars:      make(map[string]string),
		}
		for sidecar := range job.Sidecars {
			statusJobInfo.Sidecars[sidecar] = containerBuilds[kubeutil.UserSidecarContainerName(sidecar)]
		}
		jobs[path] = statusJobInfo
	}

	// If we haven't logged a start timestamp yet, use now.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateRunning,
		"",
		startTimestamp,
		nil,
		services,
		jobs,
		containerBuildStatuses,
	)
	return err
}

func (c *Controller) getContainerBuilds(
	build *latticev1.Build,
	path tree.Path,
	containers map[string]definitionv1.Container,
	containerBuilds map[string]string,
	containerBuildHashes map[string]*latticev1.ContainerBuild,
	containerBuildStatuses map[string]latticev1.ContainerBuildStatus,
) error {
	for containerName, container := range containers {
		buildDefinition, err := c.hydrateContainerBuild(build, path, container.Build)
		if err != nil {
			return err
		}

		definitionHash, err := hashContainerBuild(buildDefinition)
		if err != nil {
			return err
		}

		// check if we've already observed a container build with this hash
		containerBuild, ok := containerBuildHashes[definitionHash]
		if !ok {
			// if not, see if an active or succeeded container build already exists with this hash
			containerBuild, err = c.findContainerBuildForDefinitionHash(build.Namespace, definitionHash)
			if err != nil {
				return err
			}
		}

		// if we found a container build, add an owner reference to it
		if containerBuild != nil {
			containerBuild, err := c.addOwnerReference(build, containerBuild)
			if err != nil {
				return err
			}

			containerBuildStatuses[containerBuild.Name] = containerBuild.Status
			containerBuildHashes[definitionHash] = containerBuild
			containerBuilds[containerName] = containerBuild.Name
			continue
		}

		// haven't found a container build so need to create one
		containerBuild, err = c.createNewContainerBuild(build, buildDefinition, definitionHash)
		if err != nil {
			return err
		}

		containerBuildStatuses[containerBuild.Name] = containerBuild.Status
		containerBuildHashes[definitionHash] = containerBuild
		containerBuilds[containerName] = containerBuild.Name
	}

	return nil
}

func (c *Controller) hydrateContainerBuild(
	build *latticev1.Build,
	path tree.Path,
	containerBuild *definitionv1.ContainerBuild,
) (*definitionv1.ContainerBuild, error) {
	// If the container build is a command build and a source wasn't specified,
	// use the git repository commit context that the definition was resolved from.
	if containerBuild.CommandBuild == nil || containerBuild.CommandBuild.Source != nil {
		return containerBuild, nil
	}

	i, ok := build.Spec.ResolutionInfo[path]
	if !ok {
		err := fmt.Errorf(
			"%v resolution info did not have information for %v",
			build.Description(c.namespacePrefix),
			path.String(),
		)
		return nil, err
	}

	// Copy so we don't mutate the cache
	b := &definitionv1.ContainerBuild{}
	*b = *containerBuild

	b.CommandBuild.Source = &definitionv1.ContainerBuildSource{
		GitRepository: &definitionv1.GitRepository{
			URL:    i.Commit.RepositoryURL,
			Commit: &i.Commit.Commit,
		},
	}

	if i.SSHKeySecret != nil {
		b.CommandBuild.Source.GitRepository.SSHKey = &definitionv1.SecretRef{
			Value: *i.SSHKeySecret,
		}
	}

	return b, nil
}

func hashContainerBuild(containerBuild *definitionv1.ContainerBuild) (string, error) {
	// Note: json marshalling is deterministic: https://godoc.org/encoding/json#Marshal
	// "Map values encode as JSON objects. The map's key type must either be a string,
	//  an integer type, or implement encoding.TextMarshaler. The map keys are sorted
	//  and used as JSON object keys..."
	definitionJSON, err := json.Marshal(containerBuild)
	if err != nil {
		return "", err
	}

	return sha1.EncodeToHexString(definitionJSON)
}
