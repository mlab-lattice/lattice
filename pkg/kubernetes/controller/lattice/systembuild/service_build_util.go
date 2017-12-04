package systembuild

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (sbc *Controller) getServiceBuildState(namespace, svcBuildName string) *crv1.ServiceBuildState {
	svcBuildKey := fmt.Sprintf("%v/%v", namespace, svcBuildName)
	svcBuildObj, exists, err := sbc.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil || !exists {
		return nil
	}

	return &(svcBuildObj.(*crv1.ServiceBuild).Status.State)
}

func (sbc *Controller) getServiceBuildFromInfo(svcbInfo *crv1.SystemBuildServicesInfo, ns string) (*crv1.ServiceBuild, bool, error) {
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

func (sbc *Controller) createServiceBuild(sysb *crv1.SystemBuild, svcDef *definition.Service) (*crv1.ServiceBuild, error) {
	svcBuild := getNewServiceBuildFromDefinition(sysb, svcDef)
	return sbc.latticeClient.V1().ServiceBuilds(sysb.Namespace).Create(svcBuild)
}

func getNewServiceBuildFromDefinition(sysBuild *crv1.SystemBuild, svcDefinition *definition.Service) *crv1.ServiceBuild {
	labels := map[string]string{}

	sysBuildVersionLabel, ok := sysBuild.Labels[constants.LabelKeySystemBuildVersion]
	if ok {
		labels[constants.LabelKeySystemBuildVersion] = sysBuildVersionLabel
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
			Name:            uuid.NewV4().String(),
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
