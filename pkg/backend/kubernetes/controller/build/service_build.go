package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (c *Controller) findServiceBuildForDefinitionHash(namespace, definitionHash string) (*latticev1.ServiceBuild, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceBuildDefinitionHashLabelKey, selection.Equals, []string{definitionHash})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	builds, err := c.serviceBuildLister.List(selector)
	if err != nil {
		return nil, err
	}
	for _, build := range builds {
		if build.Status.State != latticev1.ServiceBuildStateFailed {
			return build, nil
		}
	}

	return nil, nil
}

func (c *Controller) createNewServiceBuild(
	build *latticev1.Build,
	serviceDefinition definition.Service,
	definitionHash string,
) (*latticev1.ServiceBuild, error) {
	serviceBuild := serviceBuild(build, serviceDefinition, definitionHash)
	serviceBuild, err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Create(serviceBuild)
	if err != nil {
		return nil, fmt.Errorf("error creating service build for %v with definition hash %v", build.Description(c.namespacePrefix), definitionHash)
	}

	return serviceBuild, nil
}

func serviceBuild(
	build *latticev1.Build,
	serviceDefinition definition.Service,
	definitionHash string,
) *latticev1.ServiceBuild {
	spec := serviceBuildSpec(serviceDefinition)
	return &latticev1.ServiceBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
			Labels: map[string]string{
				latticev1.ServiceBuildDefinitionHashLabelKey: definitionHash,
			},
			OwnerReferences: []metav1.OwnerReference{*newOwnerReference(build)},
		},
		Spec: spec,
	}
}

func serviceBuildSpec(serviceDefinition definition.Service) latticev1.ServiceBuildSpec {
	components := map[string]latticev1.ServiceBuildSpecComponentBuildInfo{}
	for _, component := range serviceDefinition.Components() {
		components[component.Name] = latticev1.ServiceBuildSpecComponentBuildInfo{
			DefinitionBlock: component.Build,
		}
	}

	return latticev1.ServiceBuildSpec{
		Components: components,
	}
}

func (c *Controller) addOwnerReference(build *latticev1.Build, serviceBuild *latticev1.ServiceBuild) (*latticev1.ServiceBuild, error) {
	ownerRef := kubeutil.GetOwnerReference(serviceBuild, build)

	// already has the service build as an owner
	if ownerRef != nil {
		return serviceBuild, nil
	}

	// Copy so we don't mutate the cache
	serviceBuild = serviceBuild.DeepCopy()
	serviceBuild.OwnerReferences = append(serviceBuild.OwnerReferences, *newOwnerReference(build))

	serviceBuild, err := c.latticeClient.LatticeV1().ServiceBuilds(serviceBuild.Namespace).Update(serviceBuild)
	if err != nil {
		err = fmt.Errorf(
			"error adding owner reference to %v to %v: %v",
			build.Description(c.namespacePrefix),
			serviceBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return serviceBuild, nil
}

func (c *Controller) removeOwnerReference(build *latticev1.Build, serviceBuild *latticev1.ServiceBuild) (*latticev1.ServiceBuild, error) {
	found := false
	var ownerRefs []metav1.OwnerReference
	for _, ref := range serviceBuild.GetOwnerReferences() {
		if ref.UID == build.GetUID() {
			found = true
			break
		}

		ownerRefs = append(ownerRefs, ref)
	}

	if !found {
		return serviceBuild, nil
	}

	// Copy so we don't mutate the cache
	serviceBuild = serviceBuild.DeepCopy()
	serviceBuild.OwnerReferences = ownerRefs

	serviceBuild, err := c.latticeClient.LatticeV1().ServiceBuilds(serviceBuild.Namespace).Update(serviceBuild)
	if err != nil {
		err = fmt.Errorf(
			"error removing owner reference to %v from %v: %v",
			build.Description(c.namespacePrefix),
			serviceBuild.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return serviceBuild, nil
}

func newOwnerReference(build *latticev1.Build) *metav1.OwnerReference {
	gvk := latticev1.BuildKind

	// we don't want the existence of the service build to prevent the
	// build from being deleted.
	// we'll add a finalizer which removes the owner reference. once
	// the owner reference has been removed, the service build can
	// check to see if it has any owner reference still, and if not
	// it can be garbage collected.
	// FIXME: figure out what we want our build garbage collection story to be
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
