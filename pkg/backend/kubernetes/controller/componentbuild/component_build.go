package componentbuild

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) deleteComponentBuild(build *latticev1.ComponentBuild) error {
	// want to keep the component build object around until its dependents
	// have been garbage collected for bookkeeping
	foregroundDelete := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &foregroundDelete,
	}

	err := c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Delete(build.Name, deleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error deleting %v: %v", build.Description(c.namespacePrefix), err)
	}

	return nil
}

func isOrphaned(build *latticev1.ComponentBuild) bool {
	return len(build.OwnerReferences) == 0
}
