package build

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncMissingContainerBuildsBuild(build *latticev1.Build, stateInfo stateInfo) error {
	containerBuildStatuses := stateInfo.containerBuildStatuses
	containerBuildHashes := make(map[string]*latticev1.ContainerBuild)
	services := make(map[tree.NodePath]latticev1.BuildStatusService)

	// look through all the containers of each service to see if there are any containers that
	// don't have builds yet
	for _, path := range stateInfo.servicesNeedNewContainerBuilds {
		serviceInfo := build.Spec.Services[path]

		containers := map[string]definitionv1.Container{
			kubeutil.UserMainContainerName: serviceInfo.Definition.Container,
		}
		for name, sidecarContainer := range serviceInfo.Definition.Sidecars {
			containers[kubeutil.UserSidecarContainerName(name)] = sidecarContainer
		}

		// maps the service's container names to the container builds for them
		containerBuilds := make(map[string]string)
		for containerName, container := range containers {
			definitionHash, err := hashContainerBuild(container)
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
			containerBuild, err = c.createNewContainerBuild(build, container.Build, definitionHash)
			if err != nil {
				return err
			}

			containerBuildStatuses[containerBuild.Name] = containerBuild.Status
			containerBuildHashes[definitionHash] = containerBuild
			containerBuilds[containerName] = containerBuild.Name
		}

		statusServiceInfo := latticev1.BuildStatusService{
			MainContainer: containerBuilds[kubeutil.UserMainContainerName],
			Sidecars:      make(map[string]string),
		}
		for sidecar := range serviceInfo.Definition.Sidecars {
			statusServiceInfo.Sidecars[sidecar] = containerBuilds[kubeutil.UserSidecarContainerName(sidecar)]
		}
		services[path] = statusServiceInfo
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
		containerBuildStatuses,
	)
	return err
}

func hashContainerBuild(container definitionv1.Container) (string, error) {
	// Note: json marshalling is deterministic: https://godoc.org/encoding/json#Marshal
	// "Map values encode as JSON objects. The map's key type must either be a string,
	//  an integer type, or implement encoding.TextMarshaler. The map keys are sorted
	//  and used as JSON object keys..."
	definitionJSON, err := json.Marshal(container.Build)
	if err != nil {
		return "", err
	}

	// using sha1 for now. sha256 requires 64 bytes and label values can only be
	// up to 63 characters
	h := sha1.New()
	if _, err = h.Write(definitionJSON); err != nil {
		return "", err
	}

	definitionHash := hex.EncodeToString(h.Sum(nil))
	return definitionHash, nil
}
