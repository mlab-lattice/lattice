package build

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncFailedBuild(build *latticev1.Build, stateInfo stateInfo) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	var failedServices []tree.NodePath
	for service := range stateInfo.failedServiceBuilds {
		failedServices = append(failedServices, service)
	}

	sort.Slice(failedServices, func(i, j int) bool {
		return string(failedServices[i]) < string(failedServices[j])
	})

	message := "The following services failed to build:"
	for i, service := range failedServices {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + string(service)
	}

	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateFailed,
		message,
		build.Status.StartTimestamp,
		completionTimestamp,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningBuild(build *latticev1.Build, stateInfo stateInfo) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	var activeServices []tree.NodePath
	for service := range stateInfo.activeServiceBuilds {
		activeServices = append(activeServices, service)
	}

	sort.Slice(activeServices, func(i, j int) bool {
		return string(activeServices[i]) < string(activeServices[j])
	})

	message := "The following services are still building:"
	for i, service := range activeServices {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + string(service)
	}

	// If we haven't logged a start timestamp yet, use now.
	// This should only happen if we created all of the service builds
	// but then failed to update the status.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		message,
		build.Status.StartTimestamp,
		nil,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingServiceBuildsBuild(build *latticev1.Build, stateInfo stateInfo) error {
	serviceBuilds := stateInfo.serviceBuilds
	if serviceBuilds == nil {
		serviceBuilds = map[tree.NodePath]string{}
	}

	serviceBuildStatuses := stateInfo.serviceBuildStatuses
	if serviceBuildStatuses == nil {
		serviceBuildStatuses = map[string]latticev1.ServiceBuildStatus{}
	}

	for _, path := range stateInfo.needsNewServiceBuilds {
		serviceInfo := build.Spec.Services[path]

		// Note: json marshalling is deterministic: https://godoc.org/encoding/json#Marshal
		// "Map values encode as JSON objects. The map's key type must either be a string,
		//  an integer type, or implement encoding.TextMarshaler. The map keys are sorted
		//  and used as JSON object keys..."
		definitionJSON, err := json.Marshal(serviceInfo.Definition)
		if err != nil {
			return err
		}

		// using sha1 for now. sha256 requires 64 bytes and label values can only be
		// up to 63 characters
		h := sha1.New()
		if _, err = h.Write(definitionJSON); err != nil {
			return err
		}

		definitionHash := hex.EncodeToString(h.Sum(nil))

		serviceBuild, err := c.findServiceBuildForDefinitionHash(build.Namespace, definitionHash)
		if err != nil {
			return err
		}

		// found an existing service build
		if serviceBuild != nil {
			glog.V(4).Infof(
				"found %v for service %v of %v",
				serviceBuild.Description(c.namespacePrefix),
				path.String(),
				build.Description(c.namespacePrefix),
			)

			serviceBuild, err := c.addOwnerReference(build, serviceBuild)
			if err != nil {
				return err
			}

			serviceBuilds[path] = serviceBuild.Name
			serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status
			continue
		}

		// previous service build failed or does not exist
		// create a new one
		glog.V(4).Infof("no service build found for path %v of %v", path.String(), build.Description(c.namespacePrefix))
		serviceBuild, err = c.createNewServiceBuild(build, serviceInfo.Definition, definitionHash)
		if err != nil {
			return err
		}

		glog.V(4).Infof(
			"created %v for service %v of %v",
			serviceBuild.Description(c.namespacePrefix),
			path.String(),
			build.Description(c.namespacePrefix),
		)
		serviceBuilds[path] = serviceBuild.Name
		serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status
	}

	// If we haven't logged a start timestamp yet, use now.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		"",
		startTimestamp,
		nil,
		serviceBuilds,
		serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncSucceededBuild(build *latticev1.Build, stateInfo stateInfo) error {
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateSucceeded,
		"",
		build.Status.StartTimestamp,
		build.Status.CompletionTimestamp,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) updateSystemBuildStatus(
	build *latticev1.Build,
	state latticev1.BuildState,
	message string,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	serviceBuilds map[tree.NodePath]string,
	serviceBuildStatuses map[string]latticev1.ServiceBuildStatus,
) (*latticev1.Build, error) {
	status := latticev1.BuildStatus{
		State:   state,
		Message: message,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		ServiceBuilds:        serviceBuilds,
		ServiceBuildStatuses: serviceBuildStatuses,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status

	return c.latticeClient.LatticeV1().Builds(build.Namespace).UpdateStatus(build)
}
