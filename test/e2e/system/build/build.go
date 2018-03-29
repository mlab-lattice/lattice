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
		systemID = system.CreateSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemName, systemURL)
	})

	ifSystemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list builds, but the list should be empty",
		ifSystemCreated,
		func() {
			build.List(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), nil)
		},
	)

	version := v1.SystemVersion("1.0.0")
	var buildID1 v1.BuildID
	ConditionallyIt(
		"should be able to create a build",
		ifSystemCreated,
		func() {
			buildID1 = build.Create(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), version)
		},
	)

	ifBuildCreated := If("build creation succeeded", func() bool { return systemID != "" && buildID1 != "" })
	ConditionallyIt(
		"should be able to create a build",
		ifBuildCreated,
		func() {
			build.Get(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), buildID1)
		},
	)

	ConditionallyIt(
		"should be able to list builds, and the list should only contain the created build",
		ifBuildCreated,
		func() {
			build.List(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), []v1.BuildID{buildID1})
		},
	)

	ConditionallyIt(
		"should see the build succeed",
		ifBuildCreated,
		func() {
			build.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), buildID1, 15*time.Second, 5*time.Minute)
		},
	)

	var buildID2 v1.BuildID
	ConditionallyIt(
		"should be able to build the same version again, much faster",
		ifBuildCreated,
		func() {
			buildID2 = build.BuildSuccessfully(
				context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID),
				version,
				1*time.Second,
				10*time.Second,
			)
		},
	)

	ConditionallyIt(
		"should be able to list builds, and the list should contain both builds",
		If("both builds succeeded", func() bool { return systemID != "" && buildID1 != "" && buildID2 != "" }),
		func() {
			build.List(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), []v1.BuildID{buildID1, buildID2})

		},
	)

	ConditionallyIt(
		"should be able to delete the system",
		ifSystemCreated,
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 10*time.Second)
		},
	)
})
