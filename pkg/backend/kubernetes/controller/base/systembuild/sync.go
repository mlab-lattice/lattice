package systembuild

import (
	"fmt"
	"reflect"
	"sort"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

// Warning: syncServiceBuildStates mutates svcb. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *Controller) syncServiceBuildStates(sysb *crv1.SystemBuild, info *sysBuildStateInfo) error {
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
		return fmt.Errorf("SystemBuild %v Spec.Services did not contain expected service %v", svcb.Name, service)
	}

	serviceInfo.State = &svcb.Status.State
	if serviceInfo.Components == nil {
		serviceInfo.Components = map[string]crv1.SystemBuildServicesInfoComponentInfo{}
	}

	for component, cbInfo := range svcb.Spec.Components {
		componentInfo := crv1.SystemBuildServicesInfoComponentInfo{
			Name:              cbInfo.Name,
			Status:            cbInfo.Status,
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
func (sbc *Controller) syncFailedSystemBuild(sysb *crv1.SystemBuild, failedSvcbs map[tree.NodePath]*crv1.ServiceBuild) error {
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
func (sbc *Controller) syncRunningSystemBuild(sysb *crv1.SystemBuild, activeSvcbs map[tree.NodePath]*crv1.ServiceBuild) error {
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
func (sbc *Controller) syncMissingServiceBuildsSystemBuild(sysb *crv1.SystemBuild, needsNewSvcbs []tree.NodePath) error {
	for _, service := range needsNewSvcbs {
		svcbInfo := sysb.Spec.Services[service]

		// Check if we've already created a Service. If so just grab its status.
		if svcbInfo.Name != nil {
			svcBuildState := sbc.getServiceBuildState(sysb.Namespace, *svcbInfo.Name)
			if svcBuildState == nil {
				// This shouldn't happen.
				// FIXME: send error event
				failedState := crv1.ServiceBuildStateFailed
				svcBuildState = &failedState
				//sysBuild.Spec.Services[idx].Service = &failedState
			}

			svcbInfo.State = svcBuildState
			sysb.Spec.Services[service] = svcbInfo
			continue
		}

		// Otherwise we'll have to create a new Service.
		svcb, err := sbc.createServiceBuild(sysb, &svcbInfo.Definition)
		if err != nil {
			return err
		}

		svcbInfo.Name = &(svcb.Name)
		svcbInfo.State = &(svcb.Status.State)
		sysb.Spec.Services[service] = svcbInfo
	}

	_, err := sbc.putSystemBuildUpdate(sysb)
	if err != nil {
		return err
	}

	sbc.queue.Add(fmt.Sprintf("%v/%v", sysb.Namespace, sysb.Name))
	return nil
}

func (sbc *Controller) syncSucceededSystemBuild(svcb *crv1.SystemBuild) error {
	newStatus := crv1.SystemBuildStatus{
		State: crv1.SystemBuildStateSucceeded,
	}

	_, err := sbc.putSystemBuildStatusUpdate(svcb, newStatus)
	return err
}

// Warning: putSystemBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (sbc *Controller) putSystemBuildStatusUpdate(sysb *crv1.SystemBuild, newStatus crv1.SystemBuildStatus) (*crv1.SystemBuild, error) {
	if reflect.DeepEqual(sysb.Status, newStatus) {
		return sysb, nil
	}

	sysb.Status = newStatus
	return sbc.putSystemBuildUpdate(sysb)
}

func (sbc *Controller) putSystemBuildUpdate(sysb *crv1.SystemBuild) (*crv1.SystemBuild, error) {
	return sbc.latticeClient.LatticeV1().SystemBuilds(sysb.Namespace).Update(sysb)
}