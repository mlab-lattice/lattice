package systembuild

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (sbc *Controller) getServiceBuildState(namespace, svcBuildName string) *crv1.ServiceBuildState {
	svcb, err := sbc.serviceBuildLister.ServiceBuilds(namespace).Get(svcBuildName)
	if err != nil {
		return nil
	}

	return &(svcb.Status.State)
}

func (sbc *Controller) getServiceBuildFromInfo(svcbInfo *crv1.SystemBuildServicesInfo, ns string) (*crv1.ServiceBuild, bool, error) {
	if svcbInfo.Name == nil {
		return nil, false, nil
	}

	svcb, err := sbc.serviceBuildLister.ServiceBuilds(ns).Get(*svcbInfo.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, err
		}
	}

	return svcb, true, nil
}

func (sbc *Controller) createServiceBuild(sysb *crv1.SystemBuild, svcDef *definition.Service) (*crv1.ServiceBuild, error) {
	svcBuild := getNewServiceBuildFromDefinition(sysb, svcDef)
	return sbc.latticeClient.LatticeV1().ServiceBuilds(sysb.Namespace).Create(svcBuild)
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
