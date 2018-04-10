package testingsystem

import (
	"net/http"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/test/util/lattice/v1/system"
	"github.com/mlab-lattice/lattice/test/util/versionaggregatorservice"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

const (
	V3ServiceBCVersion         = "1.0.0"
	V3ServiceAPath             = tree.NodePath("/test/a")
	V3ServiceCBPath            = tree.NodePath("/test/b/c")
	V3ServiceAPublicPort int32 = 8080
)

type V3 struct {
	systemID              v1.SystemID
	v1client              v1client.Interface
	numServiceCBInstances int32
}

func NewV3(client v1client.Interface, systemID v1.SystemID, numServiceCBInstances int32) *V3 {
	return &V3{
		systemID:              systemID,
		v1client:              client,
		numServiceCBInstances: numServiceCBInstances,
	}
}

func (v *V3) ValidateStable() {
	sys := system.Get(v.v1client.Systems(), v.systemID)

	Expect(sys.State).To(Equal(v1.SystemStateStable))

	Expect(len(sys.Services)).To(Equal(2))
	serviceCB, ok := sys.Services[V3ServiceCBPath]
	Expect(ok).To(BeTrue())
	Expect(serviceCB.State).To(Equal(v1.ServiceStateStable))
	Expect(serviceCB.StaleInstances).To(Equal(int32(0)))
	Expect(serviceCB.UpdatedInstances).To(Equal(v.numServiceCBInstances))
	Expect(len(serviceCB.PublicPorts)).To(Equal(0))

	serviceA, ok := sys.Services[V3ServiceAPath]
	Expect(ok).To(BeTrue())

	Expect(serviceA.State).To(Equal(v1.ServiceStateStable))
	Expect(serviceA.StaleInstances).To(Equal(int32(0)))
	Expect(serviceA.UpdatedInstances).To(Equal(int32(1)))
	Expect(len(serviceA.PublicPorts)).To(Equal(1))
	port, ok := serviceA.PublicPorts[V3ServiceAPublicPort]
	Expect(ok).To(BeTrue())

	err := v.poll(port.Address, time.Second, 30*time.Second)
	Expect(err).To(Not(HaveOccurred()))
}

func (v *V3) test(serviceAURL string) error {
	client := versionaggregatorservice.NewClient(serviceAURL)
	aServiceURL := "http://a.test.local:8080"
	bCServiceURL := "http://c.b.test.local:8080"
	statusOK := http.StatusOK

	return client.CheckStatusAndAggregation(
		[]versionaggregatorservice.VersionService{{URL: bCServiceURL}},
		[]versionaggregatorservice.VersionAggregatorService{
			{
				URL: aServiceURL,
				RequestBody: &versionaggregatorservice.RequestBody{
					VersionServices: []versionaggregatorservice.VersionService{
						{URL: bCServiceURL},
					},
				},
			},
		},
		&versionaggregatorservice.Aggregation{
			VersionServices: []versionaggregatorservice.VersionServiceResponseInfo{
				{
					URL:    bCServiceURL,
					Status: &statusOK,
					Body: &versionaggregatorservice.VersionServiceResponseBody{
						Version: V3ServiceBCVersion,
					},
				},
			},
			VersionAggregatorServices: []versionaggregatorservice.VersionAggregatorServiceResponseInfo{
				{
					URL:    aServiceURL,
					Status: &statusOK,
					Body: &versionaggregatorservice.Aggregation{
						VersionServices: []versionaggregatorservice.VersionServiceResponseInfo{
							{
								URL:    bCServiceURL,
								Status: &statusOK,
								Body: &versionaggregatorservice.VersionServiceResponseBody{
									Version: V3ServiceBCVersion,
								},
							},
						},
						VersionAggregatorServices: []versionaggregatorservice.VersionAggregatorServiceResponseInfo{},
					},
				},
			},
		},
	)
}

func (v *V3) poll(serviceAURL string, interval, timeout time.Duration) error {
	err := wait.Poll(interval, timeout, func() (bool, error) {
		return false, v.test(serviceAURL)
	})
	if err == nil || err == wait.ErrWaitTimeout {
		return nil
	}

	return err
}
