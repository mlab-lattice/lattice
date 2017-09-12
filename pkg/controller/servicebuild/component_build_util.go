package servicebuild

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func getComponentBuildDefinitionHashFromLabel(cBuild *crv1.ComponentBuild) *string {
	cBuildHashLabel, ok := cBuild.Labels[componentBuildDefinitionHashLabelName]
	if !ok {
		return nil
	}
	return &cBuildHashLabel
}

func (sbc *ServiceBuildController) getComponentBuildFromInfo(cBuildInfo *crv1.ServiceBuildComponentBuildInfo, namespace string) (*crv1.ComponentBuild, bool, error) {
	if cBuildInfo.Name == nil {
		return nil, false, fmt.Errorf("ComponentBuildInfo does not contain Name")
	}

	cBuildKey := fmt.Sprintf("%v/%v", namespace, *cBuildInfo.Name)
	cBuildObj, exists, err := sbc.componentBuildStore.GetByKey(cBuildKey)
	if err != nil || !exists {
		return nil, false, err
	}

	return cBuildObj.(*crv1.ComponentBuild), true, nil
}

func getNewComponentBuildFromInfo(cBuildInfo *crv1.ServiceBuildComponentBuildInfo, namespace string) *crv1.ComponentBuild {
	labels := map[string]string{
		componentBuildDefinitionHashLabelName: *cBuildInfo.DefinitionHash,
	}
	return &crv1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
			Name:   string(uuid.NewUUID()),
		},
		Spec: crv1.ComponentBuildSpec{
			BuildDefinitionBlock: cBuildInfo.DefinitionBlock,
		},
		Status: crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStatePending,
		},
	}
}
