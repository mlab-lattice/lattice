package system

import (
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncDeletingSystem(system *latticev1.System) error {
	systemNamespace := kubernetes.SystemNamespace(c.latticeID, types.SystemID(system.Name))
	ns, err := c.kubeClient.CoreV1().Namespaces().Get(systemNamespace, metav1.GetOptions{})
	if err != nil {
		// If the namespace has been fully terminated, the system can be deleted as well,
		// so remove the finalizer.
		if errors.IsNotFound(err) {
			_, err := c.removeFinalizer(system)
			return err
		}

		return err
	}

	// Have already deleted the namespace, so waiting for it to finish terminating.
	// The system should be requeued when the namespace changes.
	if ns.DeletionTimestamp != nil {
		return nil
	}

	// Delete the namespace. This will put the namespace into the "terminating" phase.
	// Once it goes out of the terminating phase, the system should be requeued
	// and the namespace will not be found, resulting in the finalizer being removed
	// and the system being fully deleted.
	return c.kubeClient.CoreV1().Namespaces().Delete(ns.Name, nil)
}
