package system

import (
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

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
) *crv1.Service {
	labels := map[string]string{}

	sysVersionLabel, ok := sys.Labels[crv1.SystemVersionLabelKey]
	if ok {
		labels[crv1.SystemVersionLabelKey] = sysVersionLabel
	} else {
		// FIXME: add warn event
	}

	return &crv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            string(uuid.NewUUID()),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(sys, controllerKind)},
		},
		Spec: crv1.ServiceSpec{
			Path:       svcPath,
			Definition: *svcDefinition,
			BuildName:  svcBuildName,
		},
		Status: crv1.ServiceStatus{
			State: crv1.ServiceStateRollingOut,
		},
	}
}
