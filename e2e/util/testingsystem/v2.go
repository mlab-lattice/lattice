package testingsystem

import (
	"time"

	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system"
	"github.com/mlab-lattice/lattice/e2e/util/versionservice"
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

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
	Expect(len(service.Ports)).To(Equal(1))
	address, ok := service.Ports[V2ServiceAPublicPort]
	Expect(ok).To(BeTrue())

	err := v.poll(address, time.Second, 10*time.Second)
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
