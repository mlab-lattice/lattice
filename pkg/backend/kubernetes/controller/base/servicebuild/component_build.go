package servicebuild

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/satori/go.uuid"
)

func getComponentBuildDefinitionHashFromLabel(componentBuild *latticev1.ComponentBuild) *string {
	cBuildHashLabel, ok := componentBuild.Annotations[constants.AnnotationKeyComponentBuildDefinitionHash]
	if !ok {
		return nil
	}
	return &cBuildHashLabel
}

func (c *Controller) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*latticev1.ComponentBuild, error) {
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

		if *cbHashLabel == definitionHash && cb.Status.State != latticev1.ComponentBuildStateFailed {
			return cb, nil
		}
	}

	return nil, nil
}

func (c *Controller) createNewComponentBuild(
	namespace string,
	componentBuildInfo latticev1.ServiceBuildSpecComponentBuildInfo,
	definitionHash string,
	previousCbName *string,
) (*latticev1.ComponentBuild, error) {
	// If there is no new entry in the build cache, create a new ComponentBuild.
	componentBuild := newComponentBuild(componentBuildInfo, definitionHash)
	componentBuild, err := c.latticeClient.LatticeV1().ComponentBuilds(namespace).Create(componentBuild)
	if err != nil {
		return nil, err
	}

	return componentBuild, nil
}

func newComponentBuild(cbInfo latticev1.ServiceBuildSpecComponentBuildInfo, definitionHash string) *latticev1.ComponentBuild {
	cbAnnotations := map[string]string{
		constants.AnnotationKeyComponentBuildDefinitionHash: definitionHash,
	}

	return &latticev1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: cbAnnotations,
			Name:        uuid.NewV4().String(),
		},
		Spec: latticev1.ComponentBuildSpec{
			BuildDefinitionBlock: cbInfo.DefinitionBlock,
		},
		Status: latticev1.ComponentBuildStatus{
			State: latticev1.ComponentBuildStatePending,
		},
	}
}
