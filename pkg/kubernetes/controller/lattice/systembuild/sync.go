package systembuild

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

// Warning: syncServiceBuildStates mutates svcb. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *SystemBuildController) syncServiceBuildStates(sysb *crv1.SystemBuild, info *sysBuildStateInfo) error {
	for service, svcb := range info.successfulSvcbs {
		err := updateServiceBuildInfoState(sysb, service, svcb)
		if err != nil {
			return err
		}
	}

	for service, svcb := range info.activeSvcbs {
		err := updateServiceBuildInfoState(sysb, service, svcb)
		if err != nil {
			return err
		}
	}

	for service, svcb := range info.failedSvcbs {
		err := updateServiceBuildInfoState(sysb, service, svcb)
		if err != nil {
			return err
		}
	}

	result, err := sbc.putSystemBuildUpdate(sysb)
	if err != nil {
		return err
	}
	*sysb = *result
	return nil
}

func updateServiceBuildInfoState(sysb *crv1.SystemBuild, service tree.NodePath, svcb *crv1.ServiceBuild) error {
	serviceInfo, ok := sysb.Spec.Services[service]
	if !ok {
		return fmt.Errorf("SystemBuild %v Spec.Services did not contain expected service", svcb.Name, service)
	}

	serviceInfo.BuildState = &svcb.Status.State
	if serviceInfo.Components == nil {
		serviceInfo.Components = map[string]crv1.SystemBuildServicesInfoComponentInfo{}
	}

	for component, cbInfo := range svcb.Spec.Components {
		componentInfo := crv1.SystemBuildServicesInfoComponentInfo{
			BuildName:         cbInfo.BuildName,
			BuildState:        cbInfo.BuildState,
			LastObservedPhase: cbInfo.LastObservedPhase,
			FailureInfo:       cbInfo.FailureInfo,
		}
		serviceInfo.Components[component] = componentInfo
	}

	sysb.Spec.Services[service] = serviceInfo
	return nil
}

// Warning: syncFailedServiceBuild mutates sysb. Do not pass in a pointer to a SystemBuild
// from the shared cache.
func (sbc *SystemBuildController) syncFailedSystemBuild(sysb *crv1.SystemBuild, failedSvcbs map[tree.NodePath]*crv1.ServiceBuild) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	failedServices := []tree.NodePath{}
	for service := range failedSvcbs {
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

	newStatus := crv1.SystemBuildStatus{
		State:   crv1.SystemBuildStateFailed,
		Message: message,
	}

	_, err := sbc.putSystemBuildStatusUpdate(sysb, newStatus)
	return err
}

// Warning: syncRunningSystemBuild mutates sysb. Do not pass in a pointer to a SystemBuild
// from the shared cache.
func (sbc *SystemBuildController) syncRunningSystemBuild(sysb *crv1.SystemBuild, activeSvcbs map[tree.NodePath]*crv1.ServiceBuild) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	activeServices := []tree.NodePath{}
	for service := range activeSvcbs {
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

	newStatus := crv1.SystemBuildStatus{
		State:   crv1.SystemBuildStateRunning,
		Message: message,
	}

	_, err := sbc.putSystemBuildStatusUpdate(sysb, newStatus)
	return err
}

// Warning: syncMissingServiceBuildsSystemBuild mutates sysb. Do not pass in a pointer to a SystemBuild
// from the shared cache.
func (sbc *SystemBuildController) syncMissingServiceBuildsSystemBuild(sysb *crv1.SystemBuild, needsNewSvcbs []tree.NodePath) error {
	for _, service := range needsNewSvcbs {
		svcbInfo := sysb.Spec.Services[service]

		// Check if we've already created a Service. If so just grab its status.
		if svcbInfo.BuildName != nil {
			svcBuildState := sbc.getServiceBuildState(sysb.Namespace, *svcbInfo.BuildName)
			if svcBuildState == nil {
				// This shouldn't happen.
				// FIXME: send error event
				failedState := crv1.ServiceBuildStateFailed
				svcBuildState = &failedState
				//sysBuild.Spec.Services[idx].Service = &failedState
			}

			svcbInfo.BuildState = svcBuildState
			sysb.Spec.Services[service] = svcbInfo
			continue
		}

		// Otherwise we'll have to create a new Service.
		svcb, err := sbc.createServiceBuild(sysb, &svcbInfo.Definition)
		if err != nil {
			return err
		}

		svcbInfo.BuildName = &(svcb.Name)
		svcbInfo.BuildState = &(svcb.Status.State)
		sysb.Spec.Services[service] = svcbInfo
	}

	_, err := sbc.putSystemBuildUpdate(sysb)
	if err != nil {
		return err
	}

	sbc.queue.Add(fmt.Sprintf("%v/%v", sysb.Namespace, sysb.Name))
	return nil
}

func (sbc *SystemBuildController) syncSucceededSystemBuild(svcb *crv1.SystemBuild) error {
	newStatus := crv1.SystemBuildStatus{
		State: crv1.SystemBuildStateSucceeded,
	}

	_, err := sbc.putSystemBuildStatusUpdate(svcb, newStatus)
	return err
}

// Warning: putSystemBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *SystemBuildController) putSystemBuildStatusUpdate(sysb *crv1.SystemBuild, newStatus crv1.SystemBuildStatus) (*crv1.SystemBuild, error) {
	if reflect.DeepEqual(sysb.Status, newStatus) {
		return sysb, nil
	}

	sysb.Status = newStatus
	return sbc.putSystemBuildUpdate(sysb)
}

func (sbc *SystemBuildController) putSystemBuildUpdate(sysb *crv1.SystemBuild) (*crv1.SystemBuild, error) {
	response := &crv1.SystemBuild{}
	err := sbc.latticeResourceClient.Put().
		Namespace(sysb.Namespace).
		Resource(crv1.ResourcePluralSystemBuild).
		Name(sysb.Name).
		Body(sysb).
		Do().
		Into(response)

	return response, err
}
