package build

import (
	"fmt"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

func WaitUntilSucceeded(client v1client.SystemBuildClient, id v1.BuildID, interval, timeout time.Duration) {
	failed := v1.BuildStateFailed
	WaitUntilInState(client, id, v1.BuildStateSucceeded, &failed, interval, timeout)
}

func WaitUntilInState(client v1client.SystemBuildClient, id v1.BuildID, state v1.BuildState, failState *v1.BuildState, interval, timeout time.Duration) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		build, err := client.Get(id)
		if err != nil {
			return false, err
		}

		if failState != nil && build.State == *failState {
			return false, fmt.Errorf("build in state %v", *failState)
		}

		return build.State == state, nil
	})
	Expect(err).NotTo(HaveOccurred())
}
