package crud

import (
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/e2e/context"
	. "github.com/mlab-lattice/system/test/util/ginkgo"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("system", func() {
	It("should be able to list systems, but the list should be empty", func() {
		system.List(context.TestContext.LatticeAPIClient.V1().Systems(), nil)
	})

	systemName := v1.SystemID("e2e-system-crud-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.Create(context.TestContext.LatticeAPIClient.V1().Systems(), systemName, systemURL)
	})

	ifSystemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list systems, and there should only be the newly created system",
		ifSystemCreated,
		func() {
			system.List(
				context.TestContext.LatticeAPIClient.V1().Systems(),
				[]v1.SystemID{systemID},
			)
		},
	)

	ConditionallyIt(
		"should be able to get the newly created system by ID",
		ifSystemCreated,
		func() {
			system.Get(
				context.TestContext.LatticeAPIClient.V1().Systems(),
				systemID,
			)
		},
	)

	ConditionallyIt(
		"should see the system become stable",
		ifSystemCreated,
		func() {
			system.WaitUntilStable(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 10*time.Second)
		},
	)

	ConditionallyIt(
		"should be able to delete the newly created system by ID",
		ifSystemCreated,
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 45*time.Second)
		},
	)
})
