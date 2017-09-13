package systembuild

import (
	"fmt"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (sbc *SystemBuildController) getServiceBuildState(namespace, svcBuildName string) *crv1.ServiceBuildState {
	svcBuildKey := fmt.Sprintf("%v/%v", namespace, svcBuildName)
	svcBuildObj, exists, err := sbc.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil || !exists {
		return nil
	}

	return &(svcBuildObj.(*crv1.ServiceBuild).Status.State)
}

func getNewServiceBuildFromDefinition(sysBuild *crv1.SystemBuild, svcDefinition *systemdefinition.Service) *crv1.ServiceBuild {
	labels := map[string]string{}

	sysBuildVersionLabel, ok := sysBuild.Labels[crv1.SystemBuildVersionLabelKey]
	if ok {
		labels[crv1.SystemBuildVersionLabelKey] = sysBuildVersionLabel
	} else {
		// FIXME: add warn event
	}

	componentBuildsInfo := []crv1.ServiceBuildComponentBuildInfo{}
	for _, component := range svcDefinition.Components {
		componentBuildsInfo = append(
			componentBuildsInfo,
			crv1.ServiceBuildComponentBuildInfo{
				DefinitionBlock: component.Build,
				ComponentName:   component.Name,
			},
		)
	}

	return &crv1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            string(uuid.NewUUID()),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(sysBuild, controllerKind)},
		},
		Spec: crv1.ServiceBuildSpec{
			ComponentBuildsInfo: componentBuildsInfo,
		},
		Status: crv1.ServiceBuildStatus{
			State: crv1.ServiceBuildStatePending,
		},
	}
}
