package build

import (
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

func WaitUntilSucceeded(client v1client.BuildClient, id v1.BuildID, interval, timeout time.Duration) {
	WaitUntilInState(client, id, v1.BuildStateSucceeded, interval, timeout)
}

func WaitUntilInState(client v1client.BuildClient, id v1.BuildID, state v1.BuildState, interval, timeout time.Duration) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		build, err := client.Get(id)
		if err != nil {
			return false, err
		}

		return build.State == state, nil
	})
	Expect(err).NotTo(HaveOccurred())
}
