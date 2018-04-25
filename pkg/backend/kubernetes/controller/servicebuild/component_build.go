package servicebuild

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/satori/go.uuid"
)

func (c *Controller) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*latticev1.ComponentBuild, error) {
	// TODO: similar scalability concerns to owningServiceBuilds
	cbs, err := c.componentBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, cb := range cbs {
		hash, ok := cb.DefinitionHashAnnotation()
		if !ok {
			// FIXME: add warn event
			continue
		}

		if hash == definitionHash && cb.Status.State != latticev1.ComponentBuildStateFailed {
			return cb, nil
		}
	}

	return nil, nil
}

func (c *Controller) createNewComponentBuild(
	build *latticev1.ServiceBuild,
	componentBuildInfo latticev1.ServiceBuildSpecComponentBuildInfo,
	definitionHash string,
) (*latticev1.ComponentBuild, error) {
	// If there is no new entry in the build cache, create a new ComponentBuild.
	componentBuild := newComponentBuild(build, componentBuildInfo, definitionHash)
	componentBuild, err := c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Create(componentBuild)
	if err != nil {
		return nil, err
	}

	return componentBuild, nil
}

func newComponentBuild(build *latticev1.ServiceBuild, cbInfo latticev1.ServiceBuildSpecComponentBuildInfo, definitionHash string) *latticev1.ComponentBuild {
	cbAnnotations := map[string]string{
		latticev1.ComponentBuildDefinitionHashAnnotationKey: definitionHash,
	}

	return &latticev1.ComponentBuild{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     cbAnnotations,
			Name:            uuid.NewV4().String(),
			OwnerReferences: []metav1.OwnerReference{*newOwnerReference(build)},
		},
		Spec: latticev1.ComponentBuildSpec{
			BuildDefinitionBlock: cbInfo.DefinitionBlock,
		},
	}
}

func (c *Controller) addOwnerReference(build *latticev1.ServiceBuild, componentBuild *latticev1.ComponentBuild) (*latticev1.ComponentBuild, error) {
	ownerRef := kubeutil.GetOwnerReference(componentBuild, build)

	// already has the service build as an owner
	if ownerRef != nil {
		return componentBuild, nil
	}

	// Copy so we don't mutate the cache
	componentBuild = componentBuild.DeepCopy()
	componentBuild.OwnerReferences = append(componentBuild.OwnerReferences, *newOwnerReference(build))

	return c.latticeClient.LatticeV1().ComponentBuilds(componentBuild.Namespace).Update(componentBuild)
}

func newOwnerReference(build *latticev1.ServiceBuild) *metav1.OwnerReference {
	gvk := latticev1.ServiceBuildKind

	// we don't want the existence of the component build to prevent the
	// service build from being deleted.
	// we'll add a finalizer which removes the owner reference. once
	// the owner reference has been removed, the component build can
	// check to see if it has any owner reference still, and if not
	// it can be garbage collected.
	// FIXME: figure out what we want our build garbage collection story to be
	blockOwnerDeletion := false

	// set isController to false, since there should only be one controller
	// owning the lifecycle of the service build. since other service builds
	// may also  end up adopting the component build, we shouldn't think of
	// any given service build as the controller service build
	isController := false

	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               build.Name,
		UID:                build.UID,
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
