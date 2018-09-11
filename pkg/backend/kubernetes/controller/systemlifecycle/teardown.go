package systemlifecycle

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) updateTeardownStatus(
	teardown *latticev1.Teardown,
	state latticev1.TeardownState,
	message string,
) (*latticev1.Teardown, error) {
	status := latticev1.TeardownStatus{
		State:              state,
		ObservedGeneration: teardown.Generation,
		Message:            message,
	}

	if reflect.DeepEqual(teardown.Status, status) {
		return teardown, nil
	}

	// Copy so the shared cache isn't mutated
	teardown = teardown.DeepCopy()
	teardown.Status = status

	result, err := c.latticeClient.LatticeV1().Teardowns(teardown.Namespace).UpdateStatus(teardown)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", teardown.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) acquireTeardownLock(teardown *latticev1.Teardown) error {
	namespace, err := c.kubeNamespaceLister.Get(teardown.Namespace)
	if err != nil {
		return err
	}

	return c.lifecycleActions.AcquireTeardown(namespace.UID, teardown.V1ID())
}

func (c *Controller) releaseTeardownLock(teardown *latticev1.Teardown) error {
	namespace, err := c.kubeNamespaceLister.Get(teardown.Namespace)
	if err != nil {
		return err
	}

	c.lifecycleActions.ReleaseTeardown(namespace.UID, teardown.V1ID())
	return nil
}
