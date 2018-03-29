package deploy

import (
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

func WaitUntilSucceeded(client v1client.DeployClient, id v1.DeployID, interval, timeout time.Duration) {
	WaitUntilInState(client, id, v1.DeployStateSucceeded, interval, timeout)
}

func WaitUntilInState(client v1client.DeployClient, id v1.DeployID, state v1.DeployState, interval, timeout time.Duration) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		deploy, err := client.Get(id)
		if err != nil {
			return false, err
		}

		return deploy.State == state, nil
	})
	Expect(err).NotTo(HaveOccurred())
}
