package servicebuild

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (c *Controller) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*latticev1.ContainerBuild, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ContainerBuildDefinitionHashLabelKey, selection.Equals, []string{definitionHash})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	componentBuilds, err := c.componentBuildLister.List(selector)
	if err != nil {
		return nil, err
	}

	// look for a component build that is either running or successfully completed,
	// and is not actively being garbage collected
	for _, componentBuild := range componentBuilds {
		if componentBuild.DeletionTimestamp != nil {
			continue
		}

		if componentBuild.Status.State == latticev1.ComponentBuildStateFailed {
			continue
		}

		return componentBuild, nil
	}

	return nil, nil
}

func (c *Controller) createNewComponentBuild(
	build *latticev1.ServiceBuild,
	componentBuildInfo latticev1.ServiceBuildSpecComponentBuildInfo,
	definitionHash string,
) (*latticev1.ContainerBuild, error) {
	// If there is no new entry in the build cache, create a new ContainerBuild.
	componentBuild := newComponentBuild(build, componentBuildInfo, definitionHash)
	result, err := c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Create(componentBuild)
	if err != nil {
		return nil, fmt.Errorf("error creating component build for %v with definition hash %v", build.Description(c.namespacePrefix), definitionHash)
	}

	return result, nil
}

func newComponentBuild(build *latticev1.ServiceBuild, cbInfo latticev1.ServiceBuildSpecComponentBuildInfo, definitionHash string) *latticev1.ContainerBuild {
	return &latticev1.ContainerBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			OwnerReferences: []metav1.OwnerReference{*newOwnerReference(build)},
			Labels: map[string]string{
				latticev1.ContainerBuildDefinitionHashLabelKey: definitionHash,
			},
		},
		Spec: latticev1.ContainerBuildSpec{
			BuildDefinitionBlock: cbInfo.DefinitionBlock,
		},
	}
}

func (c *Controller) addOwnerReference(build *latticev1.ServiceBuild, componentBuild *latticev1.ContainerBuild) (*latticev1.ContainerBuild, error) {
	ownerRef := kubeutil.GetOwnerReference(componentBuild, build)

	// already has the service build as an owner
	if ownerRef != nil {
		return componentBuild, nil
	}

	// Copy so we don't mutate the cache
	componentBuild = componentBuild.DeepCopy()
	componentBuild.OwnerReferences = append(componentBuild.OwnerReferences, *newOwnerReference(build))

	result, err := c.latticeClient.LatticeV1().ComponentBuilds(componentBuild.Namespace).Update(componentBuild)
	if err != nil {
		err = fmt.Errorf(
			"error adding owner reference (owner: %v, dependent: %v): %v",
			build.Description(c.namespacePrefix),
			componentBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func (c *Controller) removeOwnerReference(build *latticev1.ServiceBuild, componentBuild *latticev1.ContainerBuild) (*latticev1.ContainerBuild, error) {
	found := false
	var ownerRefs []metav1.OwnerReference
	for _, ref := range componentBuild.GetOwnerReferences() {
		if ref.UID == build.GetUID() {
			found = true
			break
		}

		ownerRefs = append(ownerRefs, ref)
	}

	if !found {
		return componentBuild, nil
	}

	// Copy so we don't mutate the cache
	componentBuild = componentBuild.DeepCopy()
	componentBuild.OwnerReferences = ownerRefs

	result, err := c.latticeClient.LatticeV1().ComponentBuilds(componentBuild.Namespace).Update(componentBuild)
	if err != nil {
		err = fmt.Errorf(
			"error removing owner reference (owner: %v, dependent: %v): %v",
			build.Description(c.namespacePrefix),
			componentBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func newOwnerReference(build *latticev1.ServiceBuild) *metav1.OwnerReference {
	gvk := latticev1.ServiceBuildKind

	// we don't want the existence of the component build to prevent the
	// service build from being deleted.
	// we'll add a finalizer which removes the owner reference. once
	// the owner reference has been removed, the component build can
	// check to see if it has any owner reference still, and if not
	// it can be garbage collected.
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
