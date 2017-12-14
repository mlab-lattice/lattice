package servicebuild

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/satori/go.uuid"
)

func getComponentBuildDefinitionHashFromLabel(cb *crv1.ComponentBuild) *string {
	cBuildHashLabel, ok := cb.Annotations[constants.AnnotationKeyComponentBuildDefinitionHash]
	if !ok {
		return nil
	}
	return &cBuildHashLabel
}

func (sbc *Controller) getComponentBuildFromInfo(cbInfo *crv1.ServiceBuildComponentBuildInfo, ns string) (*crv1.ComponentBuild, bool, error) {
	if cbInfo.Name == nil {
		return nil, false, nil
	}

	cb, err := sbc.componentBuildLister.ComponentBuilds(ns).Get(*cbInfo.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return cb, true, nil
}

func (sbc *Controller) findComponentBuildForDefinitionHash(ns, definitionHash string) (*crv1.ComponentBuild, error) {
	// Check recent build cache
	cb, err := sbc.findComponentBuildInRecentBuildCache(ns, definitionHash)
	if err != nil {
		return nil, err
	}

	// If we found a build in the recent build cache return it.
	if cb != nil {
		return cb, nil
	}

	// We couldn't find a ComponentBuild in our recent builds cache, so we'll check the Store to see if
	// there's a ComponentBuild we could use in there.
	return sbc.findComponentBuildInStore(ns, definitionHash)
}

func (sbc *Controller) findComponentBuildInRecentBuildCache(ns, definitionHash string) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.RLock()
	defer sbc.recentComponentBuildsLock.RUnlock()

	cbNsCache, ok := sbc.recentComponentBuilds[ns]
	if !ok {
		return nil, nil
	}

	cbName, ok := cbNsCache[definitionHash]
	if !ok {
		return nil, nil
	}

	// Check to see if this build is in our ComponentBuild store
	cb, err := sbc.componentBuildLister.ComponentBuilds(ns).Get(cbName)
	if err == nil {
		return cb, nil
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}

	// The ComponentBuild isn't in our store, so we'll need to read from the API
	cb, err = sbc.latticeClient.LatticeV1().ComponentBuilds(ns).Get(cbName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// FIXME: send warn event, this shouldn't happen
			return nil, nil
		}
		return nil, err
	}

	return cb, nil
}

func (sbc *Controller) findComponentBuildInStore(ns, definitionHash string) (*crv1.ComponentBuild, error) {
	// TODO: similar scalability concerns to getServiceBuildsForComponentBuild
	cbs, err := sbc.componentBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, cb := range cbs {
		cbHashLabel := getComponentBuildDefinitionHashFromLabel(cb)
		if cbHashLabel == nil {
			// FIXME: add warn event
			continue
		}

		if *cbHashLabel == definitionHash && cb.Status.State != crv1.ComponentBuildStateFailed {
			return cb, nil
		}
	}

	return nil, nil
}

func (sbc *Controller) createComponentBuild(ns string, cbInfo *crv1.ServiceBuildComponentBuildInfo, previousCbName *string) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.Lock()
	defer sbc.recentComponentBuildsLock.Unlock()

	if cbNsCache, ok := sbc.recentComponentBuilds[ns]; ok {
		// If there is a new entry in the recent build cache, a different service build has attempted to
		// build the same component. We'll use this ComponentBuild as ours.
		cbName, ok := cbNsCache[*cbInfo.DefinitionHash]
		if ok && (previousCbName == nil || cbName != *previousCbName) {
			return sbc.getComponentBuildFromAPI(ns, cbName)
		}
	}

	// If there is no new entry in the build cache, create a new ComponentBuild.
	cb := getNewComponentBuildFromInfo(cbInfo)
	result, err := sbc.latticeClient.LatticeV1().ComponentBuilds(ns).Create(cb)
	if err != nil {
		return nil, err
	}

	cbNsCache, ok := sbc.recentComponentBuilds[ns]
	if !ok {
		cbNsCache = map[string]string{}
		sbc.recentComponentBuilds[ns] = cbNsCache
	}
	cbNsCache[*cbInfo.DefinitionHash] = cb.Name

	return result, nil
}

func (sbc *Controller) getComponentBuildFromAPI(ns, name string) (*crv1.ComponentBuild, error) {
	return sbc.latticeClient.LatticeV1().ComponentBuilds(ns).Get(name, metav1.GetOptions{})
}

func getNewComponentBuildFromInfo(cbInfo *crv1.ServiceBuildComponentBuildInfo) *crv1.ComponentBuild {
	cbAnnotations := map[string]string{
		constants.AnnotationKeyComponentBuildDefinitionHash: *cbInfo.DefinitionHash,
	}

	return &crv1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: cbAnnotations,
			Name:        uuid.NewV4().String(),
		},
		Spec: crv1.ComponentBuildSpec{
			BuildDefinitionBlock: cbInfo.DefinitionBlock,
		},
		Status: crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStatePending,
		},
	}
}
