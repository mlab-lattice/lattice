package crud

import (
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/e2e/context"
	. "github.com/mlab-lattice/system/test/util/ginkgo"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system/expected"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("system", func() {
	It("should be able to list systems, but the list should be empty", func() {
		system.List(context.TestContext.LatticeAPIClient.Systems(), nil)
	})

	systemName := v1.SystemID("e2e-system-crud-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"
	expectedSystem := &expected.System{
		ValidStates: []v1.SystemState{
			v1.SystemStatePending,
			v1.SystemStateStable,
		},
		DesiredState:    v1.SystemStateStable,
		DefinitionURL:   systemURL,
		ValidServices:   nil,
		DesiredServices: nil,
	}

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.Create(context.TestContext.LatticeAPIClient.Systems(), systemName, systemURL)
		expectedSystem.ID = systemID
	})

	createSucceeded := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list systems, and there should only be the newly created system",
		func() {
			system.List(
				context.TestContext.LatticeAPIClient.Systems(),
				[]expected.System{*expectedSystem},
			)
		},
		createSucceeded,
	)

	ConditionallyIt(
		"should be able to get the newly created system by ID",
		func() {
			system.Get(
				context.TestContext.LatticeAPIClient.Systems(),
				systemID,
				expectedSystem,
			)
		},
		createSucceeded,
	)

	ConditionallyIt(
		"should see the system become stable",
		func() {
			system.WaitUntilStable(context.TestContext.LatticeAPIClient.Systems(), systemID, 1*time.Second, 10*time.Second)
		},
		createSucceeded,
	)

	ConditionallyIt(
		"should be able to delete the newly created system by ID",
		func() {
			system.DeleteSuccesfully(context.TestContext.LatticeAPIClient.Systems(), systemID, 1*time.Second, 45*time.Second)
		},
		createSucceeded,
	)
})
