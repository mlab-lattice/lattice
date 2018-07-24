package mock

import (
	"fmt"
	"testing"
	"time"

	latticerest "github.com/mlab-lattice/lattice/pkg/api/client/rest"
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

const (
	mockSystemId     = v1.SystemID("mock-system")
	mockSystemDefURL = "https://github.com/mlab-lattice/mock-system.git"
	mockAPIServerURL = "http://localhost:8876"

	mockSystemVersion = v1.SystemVersion("1.0.0")
)

func TestMockServer(t *testing.T) {
	setupMockTest()
	t.Run("TestMockServer", mockTests)
}

func mockTests(t *testing.T) {
	latticeClient := latticerest.NewClient(mockAPIServerURL).V1()

	// test systems
	testSystems(latticeClient, t)

	// test builds
	testBuildsAndDeploys(latticeClient, t)
}

func testSystems(latticeClient v1client.Interface, t *testing.T) {
	// test create system
	fmt.Println("Test create system")
	system, err := latticeClient.Systems().Create(mockSystemId, mockSystemDefURL)
	checkErr(err, t)

	if system.ID != mockSystemId {
		t.Fatalf("Bad system returned from create")
	}
	if system.DefinitionURL != mockSystemDefURL {
		t.Fatalf("bad system url")
	}

	if system.Services != nil {
		t.Fatalf("system must not have any services yet")
	}
	fmt.Println("System created successfully")

	// test get system
	fmt.Println("Test Get System")
	s, err := latticeClient.Systems().Get(mockSystemId)
	checkErr(err, t)

	if s.ID != mockSystemId {
		t.Fatalf("Bad system returned from Get")
	}

	// test list system
	fmt.Println("Test List Systems")
	systems, err := latticeClient.Systems().List()
	checkErr(err, t)

	if len(systems) != 1 {
		t.Fatalf("Wrong number of systems")
	}

	if systems[0].ID != mockSystemId {
		t.Fatal("bad list systems")
	}
}

func testBuildsAndDeploys(latticeClient v1client.Interface, t *testing.T) {
	// create build
	fmt.Println("Test Create Build")
	build, err := latticeClient.Systems().Builds(mockSystemId).Create(mockSystemVersion)
	checkErr(err, t)

	fmt.Printf("Successfully created build. ID %v\n", build.ID)

	if build.State != v1.BuildStatePending {
		t.Fatalf("build state is not pending")
	}

	// get build
	build1, err := latticeClient.Systems().Builds(mockSystemId).Get(build.ID)
	checkErr(err, t)

	if build1 == nil {
		t.Fatal("Got build as nil")
	}

	// list builds
	builds, err := latticeClient.Systems().Builds(mockSystemId).List()

	if len(builds) != 1 {
		t.Fatal("bad # of elements for list builds")
	}

	if builds[0].ID != build.ID {
		t.Fatal("bad list builds contents")
	}

	// Deploy the build

	fmt.Printf("Depolying build %v\n", build.ID)
	deploy, err := latticeClient.Systems().Deploys(mockSystemId).CreateFromBuild(build.ID)
	checkErr(err, t)
	fmt.Printf("Created deploy %v\n", deploy.ID)

	// wait for build to run
	fmt.Printf("Waiting for build %v to enter running state\n", build.ID)
	for i := 0; i <= 20; i++ {
		build, err = latticeClient.Systems().Builds(mockSystemId).Get(build.ID)
		checkErr(err, t)
		if build.State != v1.BuildStateRunning {
			fmt.Println("...Waiting...")
			time.Sleep(200 * time.Millisecond)
		} else {
			fmt.Printf("Build %v is in running state!\n", build.ID)
			break
		}
	}

	// fail if build did not reach running state
	if build.State != v1.BuildStateRunning {
		t.Fatal("Timed out waiting for build to run")
	}

	// ensure that deploy has reached running state as well
	deploy, err = latticeClient.Systems().Deploys(mockSystemId).Get(deploy.ID)

	if deploy.State != v1.DeployStateInProgress {
		t.Fatal("Deploy must be in the `In progress` state since build is running")
	}

	// wait for build to finish
	fmt.Printf("Waiting for build %v to succeed\n", build.ID)
	for i := 0; i <= 200; i++ {
		build, err = latticeClient.Systems().Builds(mockSystemId).Get(build.ID)
		checkErr(err, t)
		if build.State != v1.BuildStateSucceeded {
			fmt.Println("...Waiting...")
			time.Sleep(200 * time.Millisecond)
		} else {
			fmt.Printf("Build %v succeeded!\n", build.ID)
			break
		}
	}

	if build.State != v1.BuildStateSucceeded {
		t.Fatal("Timed out waiting for build to succeed")
	}

	// ensure that deploy state has succeeded
	deploy, err = latticeClient.Systems().Deploys(mockSystemId).Get(deploy.ID)

	if deploy.State != v1.DeployStateSucceeded {
		t.Fatal("Deploy must be in the `succeeded` state since build has succeeded")
	}

}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
}

func setupMockTest() {
	fmt.Println("Setting up test. Starting API Server")
	// run api server
	go RunMockNewRestServer()

	fmt.Println("API server started")
}
