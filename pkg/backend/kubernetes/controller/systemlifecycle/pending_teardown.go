package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncPendingTeardown(teardown *latticev1.Teardown) error {
	err := c.acquireTeardownLock(teardown)
	if err != nil {
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			return err
		}

		_, err = c.updateTeardownStatus(
			teardown, latticev1.TeardownStateFailed,
			fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error()),
			nil,
			nil,
		)
		return err
	}

	now := metav1.Now()
	startTimestamp := &now
	completionTimestamp := &now

	_, err = c.updateTeardownStatus(
		teardown,
		latticev1.TeardownStateInProgress,
		"",
		startTimestamp,
		completionTimestamp,
	)
	return err
}
