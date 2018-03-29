package system

import (
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

func WaitUntilStable(client v1client.SystemClient, id v1.SystemID, interval, timeout time.Duration) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		system, err := client.Get(id)
		if err != nil {
			return false, err
		}

		return system.State == v1.SystemStateStable, nil
	})
	Expect(err).NotTo(HaveOccurred())
}

func WaitUntilDeleted(client v1client.SystemClient, id v1.SystemID, interval, timeout time.Duration) {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := client.Get(id)
		if err != nil {
			switch err.(type) {
			case *v1.InvalidSystemIDError:
				return true, nil
			default:
				return false, err
			}
		}

		return false, nil
	})
	Expect(err).NotTo(HaveOccurred())
}
