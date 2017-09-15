package servicebuild

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

// Warning: syncFailedServiceBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *ServiceBuildController) syncFailedServiceBuild(svcBuild *crv1.ServiceBuild, failedCBuilds []string) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same failed ComponentBuilds
	sort.Strings(failedCBuilds)

	message := "The following components failed to build:"
	for i, component := range failedCBuilds {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	newStatus := crv1.ServiceBuildStatus{
		State:   crv1.ServiceBuildStateFailed,
		Message: message,
	}

	return sbc.putServiceBuildStatusUpdate(svcBuild, newStatus)
}

func (sbc *ServiceBuildController) syncRunningServiceBuild(svcBuild *crv1.ServiceBuild, activeCBuilds []string) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same active ComponentBuilds
	sort.Strings(activeCBuilds)

	message := "The following components are still building:"
	for i, component := range activeCBuilds {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	newStatus := crv1.ServiceBuildStatus{
		State:   crv1.ServiceBuildStateRunning,
		Message: message,
	}

	return sbc.putServiceBuildStatusUpdate(svcBuild, newStatus)
}

func (sbc *ServiceBuildController) syncMissingComponentBuildsServiceBuild(svcBuild *crv1.ServiceBuild, activeCBuilds, needsNewCBuild []string) error {
	for _, component := range needsNewCBuild {
		cBuildInfo := svcBuild.Spec.ComponentBuildsInfo[component]

		// TODO: is json marshalling of a struct deterministic in order? If not could potentially get
		//		 different SHAs for the same definition. This is OK in the correctness sense, since we'll
		//		 just be duplicating work, but still not ideal
		cBuildDefinitionBlockJson, err := json.Marshal(cBuildInfo.DefinitionBlock)
		if err != nil {
			return err
		}

		h := sha256.New()
		if _, err = h.Write(cBuildDefinitionBlockJson); err != nil {
			return err
		}

		definitionHash := hex.EncodeToString(h.Sum(nil))
		cBuildInfo.DefinitionHash = &definitionHash

		cBuild, err := sbc.findComponentBuildForDefinitionHash(svcBuild.Namespace, definitionHash)
		if err != nil {
			return err
		}

		// Found an existing ComponentBuild.
		if cBuild != nil && cBuild.Status.State != crv1.ComponentBuildStateFailed {
			cBuildInfo.ComponentBuildName = &cBuild.Name
			svcBuild.Spec.ComponentBuildsInfo[component] = cBuildInfo
			continue
		}

		// Existing ComponentBuild failed. We'll try it again.
		var previousCBuildName *string
		if cBuild != nil {
			previousCBuildName = &cBuild.Name
		}

		cBuild, err = sbc.createComponentBuild(svcBuild.Namespace, &cBuildInfo, previousCBuildName)
		if err != nil {
			return err
		}

		cBuildInfo.ComponentBuildName = &cBuild.Name
		svcBuild.Spec.ComponentBuildsInfo[component] = cBuildInfo
	}

	if err := sbc.putServiceBuildUpdate(svcBuild); err != nil {
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
	activeCBuilds = append(activeCBuilds, needsNewCBuild...)
	return sbc.syncRunningServiceBuild(svcBuild, activeCBuilds)
}

func (sbc *ServiceBuildController) syncSucceededComponentBuild(svcBuild *crv1.ServiceBuild) error {
	newStatus := crv1.ServiceBuildStatus{
		State: crv1.ServiceBuildStateSucceeded,
	}

	return sbc.putServiceBuildStatusUpdate(svcBuild, newStatus)
}

// Warning: putServiceBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *ServiceBuildController) putServiceBuildStatusUpdate(svcBuild *crv1.ServiceBuild, newStatus crv1.ServiceBuildStatus) error {
	if reflect.DeepEqual(svcBuild.Status, newStatus) {
		return nil
	}

	svcBuild.Status = newStatus
	return sbc.putServiceBuildUpdate(svcBuild)
}

func (sbc *ServiceBuildController) putServiceBuildUpdate(svcBuild *crv1.ServiceBuild) error {
	err := sbc.latticeResourceRestClient.Put().
		Namespace(svcBuild.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Name(svcBuild.Name).
		Body(svcBuild).
		Do().
		Into(nil)

	return err
}
