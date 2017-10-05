package system

import (
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (sc *SystemController) getServiceState(namespace, svcName string) *crv1.ServiceState {
	svcKey := namespace + "/" + svcName
	svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
	if err != nil || !exists {
		return nil
	}

	return &(svcObj.(*crv1.Service).Status.State)
}

func getNewServiceFromDefinition(
	sys *crv1.System,
	svcDefinition *systemdefinition.Service,
	svcPath systemtree.NodePath,
	svcBuildName string,
) (*crv1.Service, error) {
	labels := map[string]string{}

	sysVersionLabel, ok := sys.Labels[crv1.SystemVersionLabelKey]
	if ok {
		labels[crv1.SystemVersionLabelKey] = sysVersionLabel
	} else {
		// FIXME: add warn event
	}

	cPortsMap := map[string][]crv1.ComponentPort{}
	ports := map[int32]bool{}
	for _, component := range svcDefinition.Components {
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

	// Need to find len(ports) + 1 unique ports to use for envoy
	// (one for ingress for each component and one for egress)
	for i := 0; i <= len(ports); i++ {

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

	if len(envoyPorts) != len(ports)+1 {
		return nil, fmt.Errorf("expected %v envoy ports but got %v", len(ports)+1, len(envoyPorts))
	}

	// Assign an envoy port to each cPort, and pop the used envoy port off the slice each time.
	for _, component := range svcDefinition.Components {
		for _, cPort := range cPortsMap[component.Name] {
			cPort.EnvoyPort = envoyPorts[0]
			envoyPorts = envoyPorts[1:]
		}
	}

	egressPort := envoyPorts[0]

	svc := &crv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            string(uuid.NewUUID()),
			Namespace:       string(sys.Spec.LatticeNamespace),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(sys, controllerKind)},
		},
		Spec: crv1.ServiceSpec{
			Path:            svcPath,
			Definition:      *svcDefinition,
			BuildName:       svcBuildName,
			EnvoyEgressPort: egressPort,
			Ports:           cPortsMap,
		},
		Status: crv1.ServiceStatus{
			State: crv1.ServiceStateRollingOut,
		},
	}

	return svc, nil
}
