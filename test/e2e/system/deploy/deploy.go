package deploy

import (
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/e2e/context"
	. "github.com/mlab-lattice/system/test/util/ginkgo"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system/build"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system/deploy"

	"fmt"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("build", func() {
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

	version := v1.SystemVersion("1.0.0")
	var buildID v1.BuildID
	ConditionallyIt(
		"should be able to create a build",
		ifSystemCreated,
		func() {
			buildID = build.BuildSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems().Builds(systemID), version, 15*time.Second, 5*time.Minute)
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

	ConditionallyIt(
		"should be able to delete the system",
		ifSystemCreated,
		func() {
			fmt.Println("about to delete system...")
			time.Sleep(2 * time.Minute)
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 2*time.Minute)
		},
	)
})
