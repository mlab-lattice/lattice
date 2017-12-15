package servicebuild

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/satori/go.uuid"
)

func getComponentBuildDefinitionHashFromLabel(componentBuild *crv1.ComponentBuild) *string {
	cBuildHashLabel, ok := componentBuild.Annotations[constants.AnnotationKeyComponentBuildDefinitionHash]
	if !ok {
		return nil
	}
	return &cBuildHashLabel
}

func (c *Controller) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	// Check recent build cache
	cb, err := c.findComponentBuildInRecentBuildCache(namespace, definitionHash)
	if err != nil {
		return nil, err
	}

	// If we found a build in the recent build cache return it.
	if cb != nil {
		return cb, nil
	}

	// We couldn't find a ComponentBuild in our recent builds cache, so we'll check the Store to see if
	// there's a ComponentBuild we could use in there.
	return c.findComponentBuildInStore(namespace, definitionHash)
}

func (c *Controller) findComponentBuildInRecentBuildCache(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	c.recentComponentBuildsLock.RLock()
	defer c.recentComponentBuildsLock.RUnlock()

	cbNsCache, ok := c.recentComponentBuilds[namespace]
	if !ok {
		return nil, nil
	}

	componentBuildName, ok := cbNsCache[definitionHash]
	if !ok {
		return nil, nil
	}

	// Check to see if this build is in our ComponentBuild store
	componentBuild, err := c.componentBuildLister.ComponentBuilds(namespace).Get(componentBuildName)
	if err == nil {
		return componentBuild, nil
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}

	// The ComponentBuild isn't in our store, so we'll need to read from the API
	componentBuild, err = c.latticeClient.LatticeV1().ComponentBuilds(namespace).Get(componentBuildName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("namespace %v had ComponentBuild %v in cache but it does not exist", namespace, componentBuildName)
		}
		return nil, err
	}

	return componentBuild, nil
}

func (c *Controller) findComponentBuildInStore(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	// TODO: similar scalability concerns to getServiceBuildsForComponentBuild
	cbs, err := c.componentBuildLister.List(labels.Everything())
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

func (c *Controller) createNewComponentBuild(
	namespace string,
	componentBuildInfo crv1.ServiceBuildSpecComponentBuildInfo,
	definitionHash string,
	previousCbName *string,
) (*crv1.ComponentBuild, error) {
	c.recentComponentBuildsLock.Lock()
	defer c.recentComponentBuildsLock.Unlock()

	if namespaceCache, ok := c.recentComponentBuilds[namespace]; ok {
		// If there is a new entry in the recent build cache, a different service build has attempted to
		// build the same component. We'll use this ComponentBuild as ours.
		componentBuildName, ok := namespaceCache[definitionHash]
		if ok && (previousCbName == nil || componentBuildName != *previousCbName) {
			componentBuild, err := c.componentBuildLister.ComponentBuilds(namespace).Get(componentBuildName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, err
				}

				componentBuild, err = c.latticeClient.LatticeV1().ComponentBuilds(namespace).Get(componentBuildName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						return nil, fmt.Errorf("namespace %v had ComponentBuild %v in cache but it does not exist", namespace, componentBuildName)
					}

					return nil, err
				}

				return componentBuild, nil
			}
		}
	}

	// If there is no new entry in the build cache, create a new ComponentBuild.
	componentBuild := newComponentBuild(componentBuildInfo, definitionHash)
	componentBuild, err := c.latticeClient.LatticeV1().ComponentBuilds(namespace).Create(componentBuild)
	if err != nil {
		return nil, err
	}

	namespaceCache, ok := c.recentComponentBuilds[namespace]
	if !ok {
		namespaceCache = map[string]string{}
		c.recentComponentBuilds[namespace] = namespaceCache
	}
	namespaceCache[definitionHash] = componentBuild.Name

	return componentBuild, nil
}

func newComponentBuild(cbInfo crv1.ServiceBuildSpecComponentBuildInfo, definitionHash string) *crv1.ComponentBuild {
	cbAnnotations := map[string]string{
		constants.AnnotationKeyComponentBuildDefinitionHash: definitionHash,
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
