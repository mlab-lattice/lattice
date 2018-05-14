package system

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) syncDeletingSystem(system *latticev1.System) error {
	// FIXME: should we teardown here or fail if not torn down here?
	systemNamespace := system.ResourceNamespace(c.namespacePrefix)
	namespace, err := c.namespaceLister.Get(systemNamespace)
	if err != nil {
		// If the namespace has been fully terminated, the system can be deleted as well,
		// so remove the finalizer.
		if errors.IsNotFound(err) {
			_, err := c.removeFinalizer(system)
			return err
		}

		return fmt.Errorf("error trying to get namespace %v for %v: %v", systemNamespace, system.Description(), err)
	}

	// Have already deleted the namespace, so waiting for it to finish terminating.
	// The system should be requeued when the namespace changes.
	if namespace.DeletionTimestamp != nil {
		return nil
	}

	// Delete the namespace. This will put the namespace into the "terminating" phase.
	// Once it goes out of the terminating phase, the system should be requeued
	// and the namespace will not be found, resulting in the finalizer being removed
	// and the system being fully deleted.
	err = c.kubeClient.CoreV1().Namespaces().Delete(namespace.Name, nil)
	if err != nil {
		return fmt.Errorf(
			"error trying to delete system %v namespace %v: %v",
			system.Description(),
			namespace.Name,
			err,
		)
	}

	return nil
}
