package servicebuild

import (
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
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

	cBuildKey := namespace + "/" + *cBuildInfo.ComponentBuildName
	cBuildObj, exists, err := sbc.componentBuildStore.GetByKey(cBuildKey)
	if err != nil || !exists {
		return nil, false, err
	}

	return cBuildObj.(*crv1.ComponentBuild), true, nil
}

func (sbc *ServiceBuildController) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	// Check recent build cache
	cBuild, err := sbc.findComponentBuildInRecentBuildCache(namespace, definitionHash)
	if err != nil {
		return nil, err
	}

	// If we found a build in the recent build cache return it.
	if cBuild != nil {
		return cBuild, nil
	}

	// We couldn't find a ComponentBuild in our recent builds cache, so we'll check the Store to see if
	// there's a ComponentBuild we could use in there.
	return sbc.findComponentBuildInStore(namespace, definitionHash), nil
}

func (sbc *ServiceBuildController) findComponentBuildInRecentBuildCache(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.RLock()
	defer sbc.recentComponentBuildsLock.RUnlock()

	cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]
	if !ok {
		return nil, nil
	}

	cBuildName, ok := cBuildNamespaceCache[definitionHash]
	if !ok {
		return nil, nil
	}

	// Check to see if this build is in our ComponentBuild store
	cBuildObj, exists, err := sbc.componentBuildStore.GetByKey(namespace + "/" + cBuildName)
	if err != nil {
		return nil, err
	}

	// The ComponentBuild exists in our store, so return that cached version.
	if exists {
		return cBuildObj.(*crv1.ComponentBuild), nil
	}

	// The ComponentBuild isn't in our store, so we'll need to read from the API
	cBuild := &crv1.ComponentBuild{}
	err = sbc.latticeResourceRestClient.Get().
		Namespace(namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cBuildName).
		Do().
		Into(cBuild)

	if err != nil {
		if errors.IsNotFound(err) {
			// FIXME: send warn event, this shouldn't happen
			return nil, nil
		}
		return nil, err
	}

	return cBuild, nil
}

func (sbc *ServiceBuildController) findComponentBuildInStore(namespace, definitionHash string) *crv1.ComponentBuild {
	// TODO: similar scalability concerns to getServiceBuildsForComponentBuild
	for _, cBuildObj := range sbc.componentBuildStore.List() {
		cBuild := cBuildObj.(*crv1.ComponentBuild)
		cBuildHashLabel := getComponentBuildDefinitionHashFromLabel(cBuild)
		if cBuildHashLabel == nil {
			// FIXME: add warn event
			continue
		}

		if *cBuildHashLabel == definitionHash && cBuild.Status.State != crv1.ComponentBuildStateFailed {
			return cBuild
		}
	}

	return nil
}

func (sbc *ServiceBuildController) createComponentBuild(
	namespace string,
	cBuildInfo *crv1.ServiceBuildComponentBuildInfo,
	previousCBuildName *string,
) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.Lock()
	defer sbc.recentComponentBuildsLock.Unlock()

	if cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]; ok {
		// If there is a new entry in the recent build cache, a different service build has attempted to
		// build the same component. We'll use this ComponentBuild as ours.
		cBuildName, ok := cBuildNamespaceCache[*cBuildInfo.DefinitionHash]
		if ok && (previousCBuildName == nil || cBuildName != *previousCBuildName) {
			return sbc.getComponentBuildFromApi(namespace, cBuildName)
		}
	}

	// If there is no new entry in the build cache, create a new ComponentBuild.
	cBuild := getNewComponentBuildFromInfo(cBuildInfo)
	result := &crv1.ComponentBuild{}
	err := sbc.latticeResourceRestClient.Post().
		Namespace(namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Body(cBuild).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]
	if !ok {
		cBuildNamespaceCache = map[string]string{}
		sbc.recentComponentBuilds[namespace] = cBuildNamespaceCache
	}
	cBuildNamespaceCache[*cBuildInfo.DefinitionHash] = cBuild.Name

	return result, nil
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
