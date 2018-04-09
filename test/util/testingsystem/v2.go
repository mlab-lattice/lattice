package testingsystem

import (
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/test/util/lattice/v1/system"
	"github.com/mlab-lattice/lattice/test/util/versionservice"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

const (
	V2ServiceAVersion          = "2.0.0"
	V2ServiceAPath             = tree.NodePath("/test/a")
	V2ServiceAPublicPort int32 = 8080
)

type V2 struct {
	systemID v1.SystemID
	v1client v1client.Interface
}

func NewV2(client v1client.Interface, systemID v1.SystemID) *V2 {
	return &V2{
		systemID: systemID,
		v1client: client,
	}
}

func (v *V2) ValidateStable() {
	sys := system.Get(v.v1client.Systems(), v.systemID)

	Expect(sys.State).To(Equal(v1.SystemStateStable))

	Expect(len(sys.Services)).To(Equal(1))
	service, ok := sys.Services[V2ServiceAPath]
	Expect(ok).To(BeTrue())

	Expect(service.State).To(Equal(v1.ServiceStateStable))
	Expect(service.StaleInstances).To(Equal(int32(0)))
	Expect(service.UpdatedInstances).To(Equal(int32(1)))
	Expect(len(service.PublicPorts)).To(Equal(1))
	port, ok := service.PublicPorts[V2ServiceAPublicPort]
	Expect(ok).To(BeTrue())

	// FIXME: remove this when terminating pod situation has been dealt with
	// The issue:
	// When updating a deployment, under the hood kubernetes creates a new replica set
	// which matches the deployment's pod template spec, and rolling replaces from
	// the old replica set to the new one.
	//
	// When a pod goes into "Terminating," it stops being included in the replica set/deployment's
	// count of pods it is responsible for.
	// When the pod goes into "Terminating," it is sent SIGTERM, and given spec.TerminationGracePeriodSeconds
	// to gracefully exit, before being sent SIGKILL.
	//
	// In this test here, version 1 of the testing system has been successfully deployed and polled.
	// Then, version 2 is rolled out (bumping the response from /test/a from 1.0.0 to 2.0.0).
	// However, the go http client keeps a connection pool with a TCP connection still alive to the original
	// pod.
	//
	// So now the pod with the new version is up and running, and the pod with the old version is in "Terminating"
	// and has been sent SIGTERM.
	// The v1 service sees that it still has an open connection though, and won't exit unless it can close that connection.
	// But since the pod is in "Terminating," the deployment does not count it as one of its pods, so the deployment
	// reports that it has 1 up to date instance and 0 stale instances, leading the system to believe it is in a stable
	// state.
	//
	// However, when the client for this test goes to make a request, it reuses its previous connection, and receives
	// 1.0.0 back as the version, even though new connections will properly connect to the new pod and get 2.0.0.
	// We should figure out if this is just acceptable behavior and document it, or if we need to work around
	// kubernetes and wait out terminating pods (could this last forever if say a node suddenly gets deleted and the pod
	// goes into Unknown?).
	//
	// For now we'll just wait TerminationGracePeriodSeconds so that the old pod will have been cleaned up and a new
	// connection will be made.
	time.Sleep(30 * time.Second)

	err := v.poll(port.Address, time.Second, 30*time.Second)
	Expect(err).To(Not(HaveOccurred()))
}

func (v *V2) test(serviceAURL string) error {
	client := versionservice.NewClient(serviceAURL)
	return client.CheckStatusAndVersion(V2ServiceAVersion)
}

func (v *V2) poll(serviceAURL string, interval, timeout time.Duration) error {
	err := wait.Poll(interval, timeout, func() (bool, error) {
		return false, v.test(serviceAURL)
	})
	if err == nil || err == wait.ErrWaitTimeout {
		return nil
	}

	return err
}
