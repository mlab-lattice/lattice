package build

import (
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/e2e/context"
	. "github.com/mlab-lattice/system/test/util/ginkgo"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system/build"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("build", func() {
	systemName := v1.SystemID("e2e-system-build-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.CreateSuccessfully(context.TestContext.LatticeAPIClient.Systems(), systemName, systemURL)
	})

	systemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list builds, but the list should be empty",
		func() {
			build.List(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), nil)
		},
		systemCreated,
	)

	version := v1.SystemVersion("1.0.0")
	var buildID1 v1.BuildID
	ConditionallyIt(
		"should be able to create a build",
		func() {
			buildID1 = build.Create(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), version)
		},
		systemCreated,
	)

	buildCreated := If("build creation succeeded", func() bool { return systemID != "" && buildID1 != "" })
	ConditionallyIt(
		"should be able to create a build",
		func() {
			build.Get(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), buildID1)
		},
		buildCreated,
	)

	ConditionallyIt(
		"should be able to list builds, and the list should only contain the created build",
		func() {
			build.List(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), []v1.BuildID{buildID1})
		},
		buildCreated,
	)

	ConditionallyIt(
		"should see the build succeed",
		func() {
			build.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), buildID1, 15*time.Second, 10*time.Minute)
		},
		buildCreated,
	)

	var buildID2 v1.BuildID
	ConditionallyIt(
		"should be able to build the same version again, much faster",
		func() {
			buildID2 = build.BuildSuccessfully(
				context.TestContext.LatticeAPIClient.Systems().Builds(systemID),
				version,
				1*time.Second,
				10*time.Second,
			)
		},
		buildCreated,
	)

	ConditionallyIt(
		"should be able to list builds, and the list should contain both builds",
		func() {
			build.List(context.TestContext.LatticeAPIClient.Systems().Builds(systemID), []v1.BuildID{buildID1, buildID2})

		},
		If("both builds succeeded", func() bool { return systemID != "" && buildID1 != "" && buildID2 != "" }),
	)

	ConditionallyIt(
		"should be able to delete the system",
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.Systems(), systemID, 1*time.Second, 10*time.Second)
		},
		systemCreated,
	)
})
