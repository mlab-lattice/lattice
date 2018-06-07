package servicebuild

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncFailedServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	// Sort the Definition names so the Status.Message is the same for the same failed ContainerBuilds
	var failedComponents []string
	for component := range stateInfo.failedComponentBuilds {
		failedComponents = append(failedComponents, component)
	}

	sort.Strings(failedComponents)

	message := "the following components failed to build:"
	for i, component := range failedComponents {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	// if we haven't logged a completion timestamp yet, use now
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateFailed,
		message,
		startTimestamp,
		completionTimestamp,
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	// Sort the Definition names so the Status.Message is the same for the same active ContainerBuilds
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

	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateRunning,
		message,
		startTimestamp,
		nil,
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingComponentBuildsServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	componentBuilds := stateInfo.componentBuilds
	if componentBuilds == nil {
		componentBuilds = make(map[string]string)
	}

	componentBuildStatuses := stateInfo.componentBuildStatuses
	if componentBuildStatuses == nil {
		componentBuildStatuses = make(map[string]latticev1.ContainerBuildStatus)
	}

	componentBuildHashes := make(map[string]*latticev1.ContainerBuild)

	for _, component := range stateInfo.needsNewComponentBuilds {
		componentInfo := build.Spec.Components[component]

		// Note: json marshalling is deterministic: https://godoc.org/encoding/json#Marshal
		// "Map values encode as JSON objects. The map's key type must either be a string,
		//  an integer type, or implement encoding.TextMarshaler. The map keys are sorted
		//  and used as JSON object keys..."
		definitionJSON, err := json.Marshal(componentInfo.DefinitionBlock)
		if err != nil {
			return err
		}

		h := sha1.New()
		if _, err = h.Write(definitionJSON); err != nil {
			return err
		}

		definitionHash := hex.EncodeToString(h.Sum(nil))

		// first check to see if we've already seen or created
		// a component build matching this hash so far
		componentBuild, ok := componentBuildHashes[definitionHash]
		if !ok {
			// if not, check all the component builds to see if one already matches the hash
			componentBuild, err = c.findComponentBuildForDefinitionHash(build.Namespace, definitionHash)
			if err != nil {
				return err
			}
		}

		// found an existing component build
		if componentBuild != nil {
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
			componentBuildHashes[definitionHash] = componentBuild
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
			"created %v for component %v of %v",
			componentBuild.Description(c.namespacePrefix),
			component,
			build.Description(c.namespacePrefix),
		)
		componentBuilds[component] = componentBuild.Name
		componentBuildStatuses[componentBuild.Name] = componentBuild.Status
		componentBuildHashes[definitionHash] = componentBuild
	}

	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateRunning,
		"",
		startTimestamp,
		nil,
		componentBuilds,
		componentBuildStatuses,
	)
	return err
}

func (c *Controller) syncSucceededServiceBuild(build *latticev1.ServiceBuild, stateInfo stateInfo) error {
	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	// if we haven't logged a completion timestamp yet, use now
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateServiceBuildStatus(
		build,
		latticev1.ServiceBuildStateSucceeded,
		"",
		startTimestamp,
		completionTimestamp,
		build.Status.ComponentBuilds,
		stateInfo.componentBuildStatuses,
	)
	return err
}
