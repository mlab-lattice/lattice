package system

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

func (sc *SystemController) getService(namespace, svcName string) (*crv1.Service, error) {
	svcKey := namespace + "/" + svcName
	svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	return svcObj.(*crv1.Service), nil
}

func (sc *SystemController) getServiceState(namespace, svcName string) (*crv1.ServiceState, error) {
	svc, err := sc.getService(namespace, svcName)
	if err != nil {
		return nil, err
	}

	if svc == nil {
		return nil, nil
	}

	return &(svc.Status.State), nil
}

func getNewService(sys *crv1.System, svcInfo *crv1.SystemServicesInfo, svcPath tree.NodePath) (*crv1.Service, error) {
	labels := map[string]string{}

	sysVersionLabel, ok := sys.Labels[constants.LabelKeySystemVersion]
	if ok {
		labels[constants.LabelKeySystemVersion] = sysVersionLabel
	} else {
		// FIXME: add warn event
	}

	spec, err := getNewServiceSpec(svcInfo, svcPath)
	if err != nil {
		return nil, err
	}

	svc := &crv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       sys.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(sys, controllerKind)},
		},
		Spec: spec,
		Status: crv1.ServiceStatus{
			State: crv1.ServiceStateRollingOut,
		},
	}

	return svc, nil
}

func getNewServiceSpec(svcInfo *crv1.SystemServicesInfo, svcPath tree.NodePath) (crv1.ServiceSpec, error) {
	cPortsMap := map[string][]crv1.ComponentPort{}
	ports := map[int32]bool{}
	for _, component := range svcInfo.Definition.Components {
		cPorts := []crv1.ComponentPort{}
		for _, port := range component.Ports {
			cPort := crv1.ComponentPort{
				Name:     port.Name,
				Port:     int32(port.Port),
				Protocol: port.Protocol,
				Public:   false,
			}
			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				cPort.Public = true
			}
			cPorts = append(cPorts, cPort)
			ports[int32(port.Port)] = true
		}

		cPortsMap[component.Name] = cPorts
	}

	var envoyPortIdx int32 = 10000
	envoyPorts := []int32{}

	// Need to find len(ports) + 2 unique ports to use for envoy
	// (one for ingress for each component, one for egress, and one for admin)
	for i := 0; i <= len(ports)+1; i++ {

		// Loop up to len(ports) + 1 times to find an unused port
		// we can use for envoy.
		for j := 0; j <= len(ports); j++ {

			// If the current envoyPortIdx is not being used by a component,
			// we'll use it for envoy. Otherwise, on to the next one.
			currPortIdx := envoyPortIdx
			envoyPortIdx += 1

			if _, ok := ports[currPortIdx]; !ok {
				envoyPorts = append(envoyPorts, currPortIdx)
				break
			}
		}
	}

	if len(envoyPorts) != len(ports)+2 {
		return crv1.ServiceSpec{}, fmt.Errorf("expected %v envoy ports but got %v", len(ports)+1, len(envoyPorts))
	}

	// Assign an envoy port to each cPort, and pop the used envoy port off the slice each time.
	for _, component := range svcInfo.Definition.Components {
		cPorts := []crv1.ComponentPort{}
		for _, cPort := range cPortsMap[component.Name] {
			cPort.EnvoyPort = envoyPorts[0]
			cPorts = append(cPorts, cPort)
			envoyPorts = envoyPorts[1:]
		}
		cPortsMap[component.Name] = cPorts
	}

	envoyAdminPort := envoyPorts[0]
	envoyEgressPort := envoyPorts[1]
	spec := crv1.ServiceSpec{
		Path:       svcPath,
		Definition: svcInfo.Definition,

		ComponentBuildArtifacts: svcInfo.ComponentBuildArtifacts,

		Ports:           cPortsMap,
		EnvoyAdminPort:  envoyAdminPort,
		EnvoyEgressPort: envoyEgressPort,
	}

	return spec, nil
}

func (sc *SystemController) createService(sys *crv1.System, svcInfo *crv1.SystemServicesInfo, svcPath tree.NodePath) (*crv1.Service, error) {
	svc, err := getNewService(sys, svcInfo, svcPath)
	if err != nil {
		return nil, err
	}

	return sc.latticeClient.V1().Services(svc.Namespace).Create(svc)
}

func (sc *SystemController) updateService(svc *crv1.Service) (*crv1.Service, error) {
	return sc.latticeClient.V1().Services(svc.Namespace).Update(svc)
}

func (sc *SystemController) deleteService(svc *crv1.Service) error {
	glog.V(5).Infof("Deleting Service %q/%q", svc.Namespace, svc.Name)
	return sc.latticeClient.V1().Services(svc.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})
}
