package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (c *Controller) findContainerBuildForDefinitionHash(namespace, definitionHash string) (*latticev1.ContainerBuild, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildDefinitionHashLabelKey, selection.Equals, []string{definitionHash})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	containerBuilds, err := c.containerBuildLister.ContainerBuilds(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// look for a service build that is either running or successfully completed,
	// and is not actively being garbage collected
	for _, containerBuild := range containerBuilds {
		if containerBuild.DeletionTimestamp != nil {
			continue
		}

		if containerBuild.Status.State == latticev1.ContainerBuildStateFailed {
			continue
		}

		return containerBuild, nil
	}

	return nil, nil
}

func (c *Controller) createNewContainerBuild(
	build *latticev1.Build,
	containerBuildDefinition *definitionv1.ContainerBuild,
	definitionHash string,
) (*latticev1.ContainerBuild, error) {
	containerBuild := containerBuild(build, containerBuildDefinition, definitionHash)
	result, err := c.latticeClient.LatticeV1().ContainerBuilds(build.Namespace).Create(containerBuild)
	if err != nil {
		return nil, fmt.Errorf("error creating service build for %v with definition hash %v", build.Description(c.namespacePrefix), definitionHash)
	}

	return result, nil
}

func containerBuild(
	build *latticev1.Build,
	containerBuildDefinition *definitionv1.ContainerBuild,
	definitionHash string,
) *latticev1.ContainerBuild {
	return &latticev1.ContainerBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			OwnerReferences: []metav1.OwnerReference{*newOwnerReference(build)},
			Labels: map[string]string{
				latticev1.ContainerBuildDefinitionHashLabelKey: definitionHash,
			},
		},
		Spec: latticev1.ContainerBuildSpec{
			Definition: containerBuildDefinition,
		},
	}
}

func (c *Controller) addOwnerReference(
	build *latticev1.Build,
	containerBuild *latticev1.ContainerBuild,
) (*latticev1.ContainerBuild, error) {
	ownerRef := kubeutil.GetOwnerReference(containerBuild, build)

	// already has the build as an owner
	if ownerRef != nil {
		return containerBuild, nil
	}

	// Copy so we don't mutate the cache
	containerBuild = containerBuild.DeepCopy()
	containerBuild.OwnerReferences = append(containerBuild.OwnerReferences, *newOwnerReference(build))

	result, err := c.latticeClient.LatticeV1().ContainerBuilds(containerBuild.Namespace).Update(containerBuild)
	if err != nil {
		err = fmt.Errorf(
			"error adding owner reference (owner: %v, dependent: %v): %v",
			build.Description(c.namespacePrefix),
			containerBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func (c *Controller) removeOwnerReference(
	build *latticev1.Build,
	containerBuild *latticev1.ContainerBuild,
) (*latticev1.ContainerBuild, error) {
	found := false
	var ownerRefs []metav1.OwnerReference
	for _, ref := range containerBuild.GetOwnerReferences() {
		if ref.UID == build.GetUID() {
			found = true
			break
		}

		ownerRefs = append(ownerRefs, ref)
	}

	if !found {
		return containerBuild, nil
	}

	// Copy so we don't mutate the cache
	containerBuild = containerBuild.DeepCopy()
	containerBuild.OwnerReferences = ownerRefs

	result, err := c.latticeClient.LatticeV1().ContainerBuilds(containerBuild.Namespace).Update(containerBuild)
	if err != nil {
		err = fmt.Errorf(
			"error removing owner reference (owner: %v, dependent: %v): %v",
			build.Description(c.namespacePrefix),
			containerBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func newOwnerReference(build *latticev1.Build) *metav1.OwnerReference {
	gvk := latticev1.BuildKind

	// we don't want the existence of the service build to prevent the
	// build from being deleted.
	// we'll add a finalizer which removes the owner reference. once
	// the owner reference has been removed, the service build can
	// check to see if it has any owner reference still, and if not
	// it can be garbage collected.
	blockOwnerDeletion := false

	// set isController to false, since there should only be one controller
	// owning the lifecycle of the service build. since other builds may also
	// end up adopting the service build, we shouldn't think of any given
	// build as the controller build
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
