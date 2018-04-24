package servicebuild

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncFailedServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same failed ComponentBuilds
	var failedComponents []string
	for component := range stateInfo.failedComponentBuilds {
		failedComponents = append(failedComponents, component)
	}

	sort.Strings(failedComponents)

	message := "The following components failed to build:"
	for i, component := range failedComponents {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateFailed,
		message,
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same active ComponentBuilds
	var activeComponents []string
	for component := range stateInfo.activeComponentBuilds {
		activeComponents = append(activeComponents, component)
	}

	sort.Strings(activeComponents)

	message := "The following components are still building:"
	for i, component := range activeComponents {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateRunning,
		message,
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingComponentBuildsServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	componentBuilds := stateInfo.componentBuilds
	if componentBuilds == nil {
		componentBuilds = map[string]string{}
	}

	componentBuildStatuses := stateInfo.componentBuildStatuses
	if componentBuildStatuses == nil {
		componentBuildStatuses = map[string]latticev1.ComponentBuildStatus{}
	}

	for _, component := range stateInfo.needsNewComponentBuilds {
		componentInfo := build.Spec.Components[component]
		definitionBlock := componentInfo.DefinitionBlock

		// FIXME: support references
		if definitionBlock.GitRepository != nil && definitionBlock.GitRepository.SSHKey != nil {
			secretName := fmt.Sprintf("%v:%v", build.Labels[constants.LabelKeyServicePath], *definitionBlock.GitRepository.SSHKey.Name)
			definitionBlock.GitRepository.SSHKey.Name = &secretName
		}

		// TODO: is json marshalling of a struct deterministic in order? If not could potentially get
		//		 different SHAs for the same definition. This is OK in the correctness sense, since we'll
		//		 just be duplicating work, but still not ideal
		definitionJSON, err := json.Marshal(componentInfo.DefinitionBlock)
		if err != nil {
			return err
		}

		h := sha256.New()
		if _, err = h.Write(definitionJSON); err != nil {
			return err
		}

		definitionHash := hex.EncodeToString(h.Sum(nil))

		componentBuild, err := c.findComponentBuildForDefinitionHash(build.Namespace, definitionHash)
		if err != nil {
			return err
		}

		// Found an existing ComponentBuild.
		if componentBuild != nil && componentBuild.Status.State != latticev1.ComponentBuildStateFailed {
			glog.V(4).Infof("Found ComponentBuild %v for %v of %v/%v", componentBuild.Name, component, build.Namespace, build.Name)

			componentBuild, err := c.addOwnerReference(build, componentBuild)
			if err != nil {
				return err
			}

			componentBuilds[component] = componentBuild.Name
			componentBuildStatuses[componentBuild.Name] = componentBuild.Status
			continue
		}

		// Existing ComponentBuild failed. We'll try it again.
		var previousCbName *string
		if componentBuild != nil {
			previousCbName = &componentBuild.Name
		}

		// TODO: there's a race here. We could create a new component build and fail before updating build.Status
		// This shouldn't actually matter in the ComponentBuild case. On the next try, the controller would find the
		// ComponentBuild thanks to the definition hash, and would use it anyways.
		glog.V(4).Infof("No ComponentBuild found for %v/%v component %v", build.Namespace, build.Name, component)
		componentBuild, err = c.createNewComponentBuild(build, componentInfo, definitionHash, previousCbName)
		if err != nil {
			return err
		}

		glog.V(4).Infof(
			"Created ComponentBuild %v for %v/%v component %v",
			componentBuild.Name,
			build.Namespace,
			build.Name,
			component,
		)
		componentBuilds[component] = componentBuild.Name
		componentBuildStatuses[componentBuild.Name] = componentBuild.Status
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateRunning,
		"",
		componentBuilds,
		componentBuildStatuses,
	)
	if err != nil {
		return err
	}

	// FIXME: ensure that these updates will create an updateServiceBuildEvent and that the ServiceBuild will be re-queued and processed again.
	// This is needed for the following scenario:
	// Service SB needs to build Component C, and finds a Running ComponentBuild CB.
	// SB decides to use it, so it will update its ComponentBuildsInfo to reflect this.
	// Before it updates however, CB finishes. When updateComponentBuild is called, SB is not found
	// as a Service to enqueue. Once SB is updated, it may never get notified that CB finishes.
	// By enqueueing it, we make sure we have up to date status information, then from there can rely
	// on updateComponentBuild to update SB's Status.
	c.queue.Add(fmt.Sprintf("%v/%v", build.Namespace, build.Name))
	return nil
}

func (c *Controller) syncSucceededServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateSucceeded,
		"",
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}

func (c *Controller) updateServiceBuildStatus(
	build *latticev1.ServiceBuild,
	state latticev1.ServiceBuildState,
	message string,
	componentBuilds map[string]string,
	componentBuildStatuses map[string]latticev1.ComponentBuildStatus,
) (*latticev1.ServiceBuild, error) {
	status := latticev1.ServiceBuildStatus{
		State:                  state,
		Message:                message,
		ComponentBuilds:        componentBuilds,
		ComponentBuildStatuses: componentBuildStatuses,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status

	return c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).UpdateStatus(build)
}
