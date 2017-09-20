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

func (sbc *SystemBuildController) getServiceBuildFromInfo(svcbInfo *crv1.SystemBuildServicesInfo, ns string) (*crv1.ServiceBuild, bool, error) {
	if svcbInfo.BuildName == nil {
		return nil, false, nil
	}

	svcbKey := ns + "/" + *svcbInfo.BuildName
	svcbObj, exists, err := sbc.serviceBuildStore.GetByKey(svcbKey)
	if err != nil || !exists {
		return nil, false, err
	}

	return svcbObj.(*crv1.ServiceBuild), true, nil
}

func (sbc *SystemBuildController) createServiceBuild(sysb *crv1.SystemBuild, svcDef *systemdefinition.Service) (*crv1.ServiceBuild, error) {
	svcBuild := getNewServiceBuildFromDefinition(sysb, svcDef)

	result := &crv1.ServiceBuild{}
	err := sbc.latticeResourceClient.Post().
		Namespace(sysb.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Body(svcBuild).
		Do().
		Into(result)
	return result, err
}

func getNewServiceBuildFromDefinition(sysBuild *crv1.SystemBuild, svcDefinition *systemdefinition.Service) *crv1.ServiceBuild {
	labels := map[string]string{}

	sysBuildVersionLabel, ok := sysBuild.Labels[crv1.SystemBuildVersionLabelKey]
	if ok {
		labels[crv1.SystemBuildVersionLabelKey] = sysBuildVersionLabel
	} else {
		// FIXME: add warn event
	}

	componentBuildsInfo := map[string]crv1.ServiceBuildComponentBuildInfo{}
	for _, component := range svcDefinition.Components {
		componentBuildsInfo[component.Name] = crv1.ServiceBuildComponentBuildInfo{
			DefinitionBlock: component.Build,
		}
	}

	return &crv1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            string(uuid.NewUUID()),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(sysBuild, controllerKind)},
		},
		Spec: crv1.ServiceBuildSpec{
			Components: componentBuildsInfo,
		},
		Status: crv1.ServiceBuildStatus{
			State: crv1.ServiceBuildStatePending,
		},
	}
}
