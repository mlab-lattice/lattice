package deploy

import (
	"time"

	"github.com/mlab-lattice/lattice/e2e/test/context"
	. "github.com/mlab-lattice/lattice/e2e/util/ginkgo"
	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system"
	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system/build"
	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system/deploy"
	"github.com/mlab-lattice/lattice/e2e/util/testingsystem"
	"github.com/mlab-lattice/lattice/pkg/api/v1"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("deploy", func() {
	//systemName := v1.SystemID("e2e-system-deploy-1")
	systemName := v1.SystemID("deploy-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.CreateSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemName, systemURL)
	})

	ifSystemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list deploys, but the list should be empty",
		ifSystemCreated,
		func() {
			deploy.List(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), nil)
		},
	)

	v1point0point0 := v1.SystemVersion("1.0.0")
	var buildID v1.BuildID
	ConditionallyIt(
		"should be able to create a build",
		ifSystemCreated,
		func() {
			buildID = build.BuildSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), v1point0point0, 15*time.Second, 5*time.Minute)
		},
	)

	ifBuildSucceeded := If("build succeeded", func() bool { return buildID != "" })
	var deployID v1.DeployID
	ConditionallyIt(
		"should be able to deploy a build",
		ifBuildSucceeded,
		func() {
			deployID = deploy.CreateFromBuild(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), buildID)
		},
	)

	ifDeployCreated := If("deploy created", func() bool { return deployID != "" })
	ConditionallyIt(
		"should be able to list deploys, and the list should only contain the created build",
		ifDeployCreated,
		func() {
			deploy.List(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), []v1.DeployID{deployID})
		},
	)

	ConditionallyIt(
		"should see the deploy succeed",
		ifDeployCreated,
		func() {
			deploy.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), deployID, 5*time.Second, 1*time.Minute)
		},
	)

	successfulV1Deploy := false
	ConditionallyIt(
		"should be able to validate the system was correctly deployed",
		ifDeployCreated,
		func() {
			v1 := testingsystem.NewV1(context.TestContext.LatticeAPIClient.V1(), systemID)
			v1.ValidateStable()
			successfulV1Deploy = true
		},
	)

	v2point0point0 := v1.SystemVersion("2.0.0")
	ifV1Deployed := If("v1 deployed successfully", func() bool { return successfulV1Deploy })
	successfulV2Deploy := false
	ConditionallyIt(
		"should be able to deploy version 2.0.0",
		ifV1Deployed,
		func() {
			deployID = deploy.CreateFromVersion(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), v2point0point0)
			deploy.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), deployID, 15*time.Second, 3*time.Minute)
			v2 := testingsystem.NewV2(context.TestContext.LatticeAPIClient.V1(), systemID)
			v2.ValidateStable()
			successfulV2Deploy = true
		},
	)

	v3point0point0 := v1.SystemVersion("3.0.0")
	ifV2Deployed := If("v2 deployed successfully", func() bool { return successfulV2Deploy })
	successfulV3point0point0Deploy := false
	ConditionallyIt(
		"should be able to deploy version 3.0.0",
		ifV2Deployed,
		func() {
			deployID := deploy.CreateFromVersion(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), v3point0point0)
			deploy.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), deployID, 15*time.Second, 3*time.Minute)
			v3 := testingsystem.NewV3(context.TestContext.LatticeAPIClient.V1(), systemID, 1)
			v3.ValidateStable()
			successfulV3point0point0Deploy = true
		},
	)

	v3point0point1 := v1.SystemVersion("3.0.1")
	ifV3point0point0Deployed := If("v2 deployed successfully", func() bool { return successfulV3point0point0Deploy })
	successfulV3point0point1Deploy := false
	ConditionallyIt(
		"should be able to deploy version 3.0.1",
		ifV3point0point0Deployed,
		func() {
			deployID := deploy.CreateFromVersion(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), v3point0point1)
			deploy.WaitUntilSucceeded(context.TestContext.LatticeAPIClient.V1().Systems().Deploys(systemID), deployID, 15*time.Second, 3*time.Minute)
			v3 := testingsystem.NewV3(context.TestContext.LatticeAPIClient.V1(), systemID, 2)
			v3.ValidateStable()
			successfulV3point0point1Deploy = true
		},
	)

	ifV3point0point1Deployed := If("v3.0.1 deployed successfully", func() bool { return successfulV3point0point1Deploy })
	ConditionallyIt(
		"should be able to delete the system",
		ifV3point0point1Deployed,
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 2*time.Minute)
		},
	)
})
