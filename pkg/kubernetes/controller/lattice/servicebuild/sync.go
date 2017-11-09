package servicebuild

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"github.com/golang/glog"
)

// Warning: syncFailedServiceBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *ServiceBuildController) syncFailedServiceBuild(svcb *crv1.ServiceBuild, failedCbs []string) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same failed ComponentBuilds
	sort.Strings(failedCbs)

	message := "The following components failed to build:"
	for i, component := range failedCbs {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	newStatus := crv1.ServiceBuildStatus{
		State:   crv1.ServiceBuildStateFailed,
		Message: message,
	}

	_, err := sbc.putServiceBuildStatusUpdate(svcb, newStatus)
	return err
}

func (sbc *ServiceBuildController) syncRunningServiceBuild(svcb *crv1.ServiceBuild, activeCbs []string) error {
	// Sort the ComponentBuild names so the Status.Message is the same for the same active ComponentBuilds
	sort.Strings(activeCbs)

	message := "The following components are still building:"
	for i, component := range activeCbs {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + component
	}

	newStatus := crv1.ServiceBuildStatus{
		State:   crv1.ServiceBuildStateRunning,
		Message: message,
	}

	_, err := sbc.putServiceBuildStatusUpdate(svcb, newStatus)
	return err
}

func (sbc *ServiceBuildController) syncMissingComponentBuildsServiceBuild(svcbs *crv1.ServiceBuild, activeCbs, needsNewCbs []string) error {
	for _, component := range needsNewCbs {
		cbInfo := svcbs.Spec.Components[component]

		// TODO: is json marshalling of a struct deterministic in order? If not could potentially get
		//		 different SHAs for the same definition. This is OK in the correctness sense, since we'll
		//		 just be duplicating work, but still not ideal
		cbDefinitionBlockJson, err := json.Marshal(cbInfo.DefinitionBlock)
		if err != nil {
			return err
		}

		h := sha256.New()
		if _, err = h.Write(cbDefinitionBlockJson); err != nil {
			return err
		}

		definitionHash := hex.EncodeToString(h.Sum(nil))
		cbInfo.DefinitionHash = &definitionHash

		cb, err := sbc.findComponentBuildForDefinitionHash(svcbs.Namespace, definitionHash)
		if err != nil {
			return err
		}

		// Found an existing ComponentBuild.
		if cb != nil && cb.Status.State != crv1.ComponentBuildStateFailed {
			glog.V(4).Infof("Found ComponentBuild %v for %v of %v", cb.Name, component, svcbs.Name)
			cbInfo.BuildName = &cb.Name
			svcbs.Spec.Components[component] = cbInfo
			continue
		}

		// Existing ComponentBuild failed. We'll try it again.
		var previousCbName *string
		if cb != nil {
			previousCbName = &cb.Name
		}

		glog.V(4).Infof("No ComponentBuild found for %v of %v", component, svcbs.Name)
		cb, err = sbc.createComponentBuild(svcbs.Namespace, &cbInfo, previousCbName)
		if err != nil {
			return err
		}

		glog.V(4).Infof("Created ComponentBuild %v for %v of %v", cb.Name, component, svcbs.Name)
		cbInfo.BuildName = &cb.Name
		svcbs.Spec.Components[component] = cbInfo
	}

	updatedSvcb, err := sbc.putServiceBuildUpdate(svcbs)
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
	activeCbs = append(activeCbs, needsNewCbs...)
	return sbc.syncRunningServiceBuild(updatedSvcb, activeCbs)
}

func (sbc *ServiceBuildController) syncSucceededServiceBuild(svcb *crv1.ServiceBuild) error {
	newStatus := crv1.ServiceBuildStatus{
		State: crv1.ServiceBuildStateSucceeded,
	}

	_, err := sbc.putServiceBuildStatusUpdate(svcb, newStatus)
	return err
}

// Warning: putServiceBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *ServiceBuildController) putServiceBuildStatusUpdate(svcb *crv1.ServiceBuild, newStatus crv1.ServiceBuildStatus) (*crv1.ServiceBuild, error) {
	if reflect.DeepEqual(svcb.Status, newStatus) {
		return svcb, nil
	}

	svcb.Status = newStatus
	return sbc.putServiceBuildUpdate(svcb)
}

func (sbc *ServiceBuildController) putServiceBuildUpdate(svcb *crv1.ServiceBuild) (*crv1.ServiceBuild, error) {
	response := &crv1.ServiceBuild{}
	err := sbc.latticeResourceClient.Put().
		Namespace(svcb.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Name(svcb.Name).
		Body(svcb).
		Do().
		Into(response)

	return response, err
}
