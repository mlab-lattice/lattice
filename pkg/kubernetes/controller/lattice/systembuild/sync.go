package systembuild

import (
	"reflect"
	"sort"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

// Warning: syncFailedServiceBuild mutates sysb. Do not pass in a pointer to a SystemBuild
// from the shared cache.
func (sbc *SystemBuildController) syncFailedSystemBuild(sysb *crv1.SystemBuild, failedSvcbs []systemtree.NodePath) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	sort.Slice(failedSvcbs, func(i, j int) bool {
		return string(failedSvcbs[i]) < string(failedSvcbs[j])
	})

	message := "The following services failed to build:"
	for i, service := range failedSvcbs {
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
func (sbc *SystemBuildController) syncRunningSystemBuild(sysb *crv1.SystemBuild, activeSvcbs []systemtree.NodePath) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	sort.Slice(activeSvcbs, func(i, j int) bool {
		return string(activeSvcbs[i]) < string(activeSvcbs[j])
	})

	message := "The following services are still building:"
	for i, service := range activeSvcbs {
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
func (sbc *SystemBuildController) syncMissingServiceBuildsSystemBuild(sysb *crv1.SystemBuild, activeSvcbs, needsNewSvcbs []systemtree.NodePath) error {
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

	updatedSysb, err := sbc.putSystemBuildUpdate(sysb)
	if err != nil {
		return err
	}

	activeSvcbs = append(activeSvcbs, needsNewSvcbs...)
	return sbc.syncRunningSystemBuild(updatedSysb, activeSvcbs)
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
		Resource(crv1.SystemBuildResourcePlural).
		Name(sysb.Name).
		Body(sysb).
		Do().
		Into(response)

	return response, err
}
