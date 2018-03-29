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
		system.List(context.TestContext.LatticeAPIClient.Systems(), nil)
	})

	systemName := v1.SystemID("e2e-system-crud-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.Create(context.TestContext.LatticeAPIClient.Systems(), systemName, systemURL)
	})

	systemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list systems, and there should only be the newly created system",
		func() {
			system.List(
				context.TestContext.LatticeAPIClient.Systems(),
				[]v1.SystemID{systemID},
			)
		},
		systemCreated,
	)

	ConditionallyIt(
		"should be able to get the newly created system by ID",
		func() {
			system.Get(
				context.TestContext.LatticeAPIClient.Systems(),
				systemID,
			)
		},
		systemCreated,
	)

	ConditionallyIt(
		"should see the system become stable",
		func() {
			system.WaitUntilStable(context.TestContext.LatticeAPIClient.Systems(), systemID, 1*time.Second, 10*time.Second)
		},
		systemCreated,
	)

	ConditionallyIt(
		"should be able to delete the newly created system by ID",
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.Systems(), systemID, 1*time.Second, 45*time.Second)
		},
		systemCreated,
	)
})
