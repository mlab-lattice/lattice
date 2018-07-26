package mock

import (
	"fmt"
	"strings"
	"testing"
	"time"

	latticerest "github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const (
	mockSystemId     = v1.SystemID("mock-system")
	mockSystemDefURL = "https://github.com/mlab-lattice/mock-system.git"
	mockAPIServerURL = "http://localhost:8876"

	mockSystemVersion = v1.SystemVersion("1.0.0")
)

var latticeClient = latticerest.NewClient(mockAPIServerURL).V1()

func TestMockServer(t *testing.T) {
	setupMockTest()
	t.Run("TestMockServer", mockTests)
}

func mockTests(t *testing.T) {

	happyPathTest(t)

}

func happyPathTest(t *testing.T) {
	createSystem(t)
	buildAndDeploy(t)
	testSecrets(t)
	checkSystemHealth(t)
	teardownSystem(t)
	deleteSystem(t)
}

func createSystem(t *testing.T) {
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

func buildAndDeploy(t *testing.T) {
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
	waitFor(func() bool {
		build, err = latticeClient.Systems().Builds(mockSystemId).Get(build.ID)
		checkErr(err, t)
		return build.State == v1.BuildStateRunning
	}, t)

	fmt.Printf("Build %v is in running state!\n", build.ID)

	// fail if build did not reach running state
	if build.State != v1.BuildStateRunning {
		t.Fatal("Timed out waiting for build to run")
	}

	// ensure that deploy has reached running state as well
	deploy, err = latticeClient.Systems().Deploys(mockSystemId).Get(deploy.ID)

	if deploy.State != v1.DeployStateInProgress {
		t.Fatal("Deploy must be in the `In progress` state since build is running")
	}

	// check service builds
	fmt.Println("Ensuring that service builds are running...")
	if build.Services == nil || len(build.Services) == 0 {
		t.Fatal("No service builds running")
	}

	// ensure that build services are running

	for _, service := range build.Services {
		if service.State != v1.ServiceBuildStateRunning {
			t.Fatal("Service build should be in running state")
		}
	}

	fmt.Println("Service builds looking good!")

	// wait for build to finish
	fmt.Printf("Waiting for build %v to succeed\n", build.ID)

	waitFor(func() bool {
		build, err = latticeClient.Systems().Builds(mockSystemId).Get(build.ID)
		checkErr(err, t)
		return build.State == v1.BuildStateSucceeded
	}, t)

	fmt.Printf("Build %v succeeded!\n", build.ID)

	// ensure that deploy state has succeeded
	fmt.Println("Verifying that deploy has succeeded...")
	deploy, err = latticeClient.Systems().Deploys(mockSystemId).Get(deploy.ID)

	if deploy.State != v1.DeployStateSucceeded {
		t.Fatal("Deploy must be in the `succeeded` state since build has succeeded")
	}
	fmt.Println("Deploy succeeded!")

	// check service builds succeeded
	fmt.Println("Ensuring that service builds has succeeded...")
	for _, service := range build.Services {
		if service.State != v1.ServiceBuildStateSucceeded {
			t.Fatal("Service build should be in running state")
		}
	}
	fmt.Println("Service builds succeeded!")

	// list deploys
	deploys, err := latticeClient.Systems().Deploys(mockSystemId).List()
	checkErr(err, t)

	if len(deploys) != 1 {
		t.Fatal("bad # of elements for list deploys")
	}

	if deploys[0].ID != deploy.ID {
		t.Fatal("bad list deploy contents")
	}
}

func testSecrets(t *testing.T) {
	fmt.Println("Testing secrets...")
	path, _ := tree.NewNodePath("/mock/test")
	fmt.Println("set secret")
	err := latticeClient.Systems().Secrets(mockSystemId).Set(path, "x", "1")
	checkErr(err, t)
	secrets, err := latticeClient.Systems().Secrets(mockSystemId).List()
	checkErr(err, t)

	fmt.Println("list secrets")
	if secrets == nil || len(secrets) != 1 {
		t.Fatal("Wrong number of secrets.")
	}

	fmt.Println("get secret")
	secret, err := latticeClient.Systems().Secrets(mockSystemId).Get(path, "x")
	checkErr(err, t)
	if secret.Value != "1" {
		t.Fatal("Bad secret.")
	}

	fmt.Println("unset secret")
	err = latticeClient.Systems().Secrets(mockSystemId).Unset(path, "x")
	checkErr(err, t)

	secrets, err = latticeClient.Systems().Secrets(mockSystemId).List()
	checkErr(err, t)

	if len(secrets) != 0 {
		t.Fatal("Secret was not unset")
	}
}

func checkSystemHealth(t *testing.T) {
	// ensure that system services are up
	system, err := latticeClient.Systems().Get(mockSystemId)
	checkErr(err, t)
	if system.Services == nil {
		t.Fatalf("system services are not set")
	}

	for _, service := range system.Services {
		if service.State != v1.ServiceStateStable {
			t.Fatalf("Service state is not stable")
		}
	}
}

func teardownSystem(t *testing.T) {

	fmt.Println("Tearing system down...")

	teardown, err := latticeClient.Systems().Teardowns(mockSystemId).Create()
	checkErr(err, t)

	fmt.Printf("Created teardown %v", teardown.ID)

	if teardown.State != v1.TeardownStatePending {
		t.Fatalf("teardown state is not pending")
	}

	// wait for teardown run
	fmt.Printf("Waiting for teardown %v to run\n", teardown.ID)

	waitFor(func() bool {
		teardown, err = latticeClient.Systems().Teardowns(mockSystemId).Get(teardown.ID)
		checkErr(err, t)
		return teardown.State == v1.TeardownStateInProgress
	}, t)

	fmt.Printf("Teardown %v entered the in progress state!\n", teardown.ID)

	// wait for teardown succeed
	fmt.Printf("Waiting for teardown %v to succeed\n", teardown.ID)

	waitFor(func() bool {
		teardown, err = latticeClient.Systems().Teardowns(mockSystemId).Get(teardown.ID)
		checkErr(err, t)
		return teardown.State == v1.TeardownStateSucceeded
	}, t)

	// check that system services are nil after teardown
	fmt.Println("Checking that system services are down after teardown...")
	system, err := latticeClient.Systems().Get(mockSystemId)
	if system.Services != nil {
		t.Fatal("System services still up")
	}

	fmt.Printf("Teardown %v succeeded!\n", teardown.ID)
}

func deleteSystem(t *testing.T) {
	system, err := latticeClient.Systems().Get(mockSystemId)
	checkErr(err, t)

	if system == nil {
		t.Fatal("System not found")
	}

	fmt.Println("Deleting system...")
	latticeClient.Systems().Delete(mockSystemId)

	_, err = latticeClient.Systems().Get(mockSystemId)

	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "invalid system") {
		t.Fatal("Expected an invalid system error")
	}
	fmt.Println("System deleted!")
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

func waitFor(condition func() bool, t *testing.T) {
	for i := 0; i <= 200; i++ {
		if !condition() {
			fmt.Println("...Waiting...")
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}

	if !condition() {
		t.Fatal("Timed out")
	}
}
