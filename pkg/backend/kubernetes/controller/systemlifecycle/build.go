package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) addBuildOwnerReference(
	deploy *latticev1.Deploy,
	build *latticev1.Build,
) (*latticev1.Build, error) {
	// check if the build already has the deploy as an owner
	ownerRef := kubeutil.GetOwnerReference(build, deploy)
	if ownerRef != nil {
		return build, nil
	}

	// Copy so we don't mutate the cache
	build = build.DeepCopy()
	build.OwnerReferences = append(build.OwnerReferences, *newOwnerReference(deploy))

	result, err := c.latticeClient.LatticeV1().Builds(build.Namespace).Update(build)
	if err != nil {
		err = fmt.Errorf(
			"error adding owner reference (owner: %v, dependent: %v): %v",
			build.Description(c.namespacePrefix),
			deploy.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func (c *Controller) removeBuildOwnerReference(
	deploy *latticev1.Deploy,
	build *latticev1.Build,
) (*latticev1.Build, error) {
	found := false
	var ownerRefs []metav1.OwnerReference
	for _, ref := range build.GetOwnerReferences() {
		if ref.UID == deploy.GetUID() {
			found = true
			break
		}

		ownerRefs = append(ownerRefs, ref)
	}

	if !found {
		return build, nil
	}

	// Copy so we don't mutate the cache
	build = build.DeepCopy()
	build.OwnerReferences = ownerRefs

	result, err := c.latticeClient.LatticeV1().Builds(build.Namespace).Update(build)
	if err != nil {
		err = fmt.Errorf(
			"error removing owner reference (owner: %v, dependent: %v): %v",
			deploy.Description(c.namespacePrefix),
			build.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func newOwnerReference(deploy *latticev1.Deploy) *metav1.OwnerReference {
	gvk := latticev1.DeployKind

	// we don't want the existence of the build to prevent the
	// deploy from being deleted.
	blockOwnerDeletion := false

	// set isController to false, since there should only be one controller
	// owning the lifecycle of the service build. since other builds may also
	// end up adopting the service build, we shouldn't think of any given
	// build as the controller build
	isController := false

	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               deploy.Name,
		UID:                deploy.UID,
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
