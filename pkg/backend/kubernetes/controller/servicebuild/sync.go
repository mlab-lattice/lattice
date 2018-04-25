package servicebuild

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

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

	message := "the following components are still building:"
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

		// found an existing component build
		if componentBuild != nil && componentBuild.Status.State != latticev1.ComponentBuildStateFailed {
			glog.V(4).Infof(
				"found %v for component %v of %v",
				componentBuild.Description(c.namespacePrefix),
				component,
				build.Description(c.namespacePrefix),
			)

			componentBuild, err := c.addOwnerReference(build, componentBuild)
			if err != nil {
				return err
			}

			componentBuilds[component] = componentBuild.Name
			componentBuildStatuses[componentBuild.Name] = componentBuild.Status
			continue
		}

		// previous component build failed or does not exist
		// create a new one
		glog.V(4).Infof("no component build found for component %v of %v", component, build.Description(c.namespacePrefix))
		componentBuild, err = c.createNewComponentBuild(build, componentInfo, definitionHash)
		if err != nil {
			return err
		}

		glog.V(4).Infof(
			"Created %v for component %v of %v",
			componentBuild.Description(c.namespacePrefix),
			component,
			build.Description(c.namespacePrefix),
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
	return err
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
