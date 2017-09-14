package servicebuild

import (
	"fmt"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const componentBuildDefinitionHashMetadataKey = "lattice-component-build-definition-hash"

func getComponentBuildDefinitionHashFromLabel(cBuild *crv1.ComponentBuild) *string {
	cBuildHashLabel, ok := cBuild.Annotations[componentBuildDefinitionHashMetadataKey]
	if !ok {
		return nil
	}
	return &cBuildHashLabel
}

func (sbc *ServiceBuildController) getComponentBuildFromInfo(
	cBuildInfo *crv1.ServiceBuildComponentBuildInfo,
	namespace string,
) (*crv1.ComponentBuild, bool, error) {
	if cBuildInfo.ComponentBuildName == nil {
		return nil, false, nil
	}

	cBuildKey := fmt.Sprintf("%v/%v", namespace, *cBuildInfo.ComponentBuildName)
	cBuildObj, exists, err := sbc.componentBuildStore.GetByKey(cBuildKey)
	if err != nil || !exists {
		return nil, false, err
	}

	return cBuildObj.(*crv1.ComponentBuild), true, nil
}

func (sbc *ServiceBuildController) getComponentBuildFromApi(namespace, name string) (*crv1.ComponentBuild, error) {
	var cBuild crv1.ComponentBuild
	err := sbc.latticeResourceRestClient.Get().
		Namespace(namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(name).
		Do().
		Into(&cBuild)
	return &cBuild, err
}

func getNewComponentBuildFromInfo(cBuildInfo *crv1.ServiceBuildComponentBuildInfo) *crv1.ComponentBuild {
	annotations := map[string]string{
		componentBuildDefinitionHashMetadataKey: *cBuildInfo.DefinitionHash,
	}
	return &crv1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Name:        string(uuid.NewUUID()),
		},
		Spec: crv1.ComponentBuildSpec{
			BuildDefinitionBlock: cBuildInfo.DefinitionBlock,
		},
		Status: crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStatePending,
		},
	}
}
