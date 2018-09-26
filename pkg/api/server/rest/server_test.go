package rest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"encoding/json"
	clientrest "github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	mockbackend "github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend"
	mockresolver "github.com/mlab-lattice/lattice/pkg/backend/mock/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	"os"
)

const (
	mockSystemID     = v1.SystemID("mock-system")
	mockRepoPath     = "/tmp/lattice/test/api/server/rest/remote"
	mockAPIServerURL = "http://localhost:8876"

	mockSystemVersion = v1.Version("1.0.0")
	mockServicePath   = tree.Path("/job")

	mockServerAPIPort = 8876
	mockServerAPIKey  = "abc"
)

var (
	latticeClient = clientrest.NewBearerTokenClient(mockAPIServerURL, mockServerAPIKey).V1()

	mockSystemDefURL = fmt.Sprintf("file://%v", mockRepoPath)

	mockSystem = &definitionv1.System{
		Description: "Mock System",
		NodePools: map[string]definitionv1.NodePool{
			"main": {
				InstanceType: "t2.small",
				NumInstances: 1,
			},
		},
		Components: map[string]definition.Component{
			"api": &definitionv1.Service{
				Description:  "api",
				NumInstances: 1,
				Container: definitionv1.Container{
					Exec: &definitionv1.ContainerExec{
						Command: []string{"foo api"},
					},
				},
			},
			"www": &definitionv1.Service{
				Description:  "www",
				NumInstances: 1,
				Container: definitionv1.Container{
					Exec: &definitionv1.ContainerExec{
						Command: []string{"foo www"},
					},
				},
			},
			"job": &definitionv1.Job{
				Description: "job",
				NodePool:    "/:main",
				Container: definitionv1.Container{
					Exec: &definitionv1.ContainerExec{
						Command: []string{"foo www"},
					},
				},
			},
		},
	}
	mockSystemBytes, _ = json.Marshal(&mockSystem)
)

func TestMockServer(t *testing.T) {
	setupMockTest()
	t.Run("TestMockServer", mockTests)
}

func mockTests(t *testing.T) {
	authTest(t)
	happyPathTest(t)
	testInvalidIDs(t)
	testInvalidDefinition(t)
}

func happyPathTest(t *testing.T) {
	createSystem(t)
	buildAndDeploy(t)
	ensureSingleDeploy(t)
	runJob(t)
	testSecrets(t)
	checkSystemHealth(t)
	teardownSystem(t)
	deleteSystem(t)
}

func createSystem(t *testing.T) {
	// test create system
	fmt.Println("Test create system")
	system, err := latticeClient.Systems().Create(mockSystemID, mockSystemDefURL)
	checkErr(err, t)

	if system.ID != mockSystemID {
		t.Fatalf("Bad system returned from create")
	}
	if system.DefinitionURL != mockSystemDefURL {
		t.Fatalf("bad system url")
	}

	fmt.Println("System created successfully")

	// test get system
	fmt.Println("Test Get System")
	s, err := latticeClient.Systems().Get(mockSystemID)
	checkErr(err, t)

	if s.ID != mockSystemID {
		t.Fatalf("Bad system returned from Get")
	}

	// test list system
	fmt.Println("Test List Systems")
	systems, err := latticeClient.Systems().List()
	checkErr(err, t)

	if len(systems) != 1 {
		t.Fatalf("Wrong number of systems")
	}

	if systems[0].ID != mockSystemID {
		t.Fatal("bad list systems")
	}

	waitFor(func() bool {
		s, err := latticeClient.Systems().Get(mockSystemID)
		checkErr(err, t)
		return s.Status.State == v1.SystemStateStable
	}, t)
}

func buildAndDeploy(t *testing.T) {
	// create build
	fmt.Println("Test Create Build")
	build, err := latticeClient.Systems().Builds(mockSystemID).CreateFromVersion(mockSystemVersion)
	checkErr(err, t)

	fmt.Printf("Successfully created build. ID %v\n", build.ID)

	if build.Status.State != v1.BuildStatePending {
		t.Fatalf("build state is not pending")
	}

	// get build
	build1, err := latticeClient.Systems().Builds(mockSystemID).Get(build.ID)
	checkErr(err, t)

	if build1 == nil {
		t.Fatal("Got build as nil")
	}

	// list builds
	builds, err := latticeClient.Systems().Builds(mockSystemID).List()

	if len(builds) != 1 {
		t.Fatal("bad # of elements for list builds")
	}

	if builds[0].ID != build.ID {
		t.Fatal("bad list builds contents")
	}

	// Deploy the build

	fmt.Printf("Depolying build %v\n", build.ID)
	deploy, err := latticeClient.Systems().Deploys(mockSystemID).CreateFromBuild(build.ID)
	checkErr(err, t)

	fmt.Printf("Created deploy %v\n", deploy.ID)

	// wait for build to run
	fmt.Printf("Waiting for build %v to enter running state\n", build.ID)
	waitFor(func() bool {
		build, err = latticeClient.Systems().Builds(mockSystemID).Get(build.ID)
		checkErr(err, t)
		return build.Status.State == v1.BuildStateRunning
	}, t)

	fmt.Printf("Build %v is in running state!\n", build.ID)

	// fail if build did not reach running state
	if build.Status.State != v1.BuildStateRunning {
		t.Fatal("Timed out waiting for build to run")
	}

	// check service builds
	fmt.Println("Ensuring that service builds are running...")
	if build.Status.Workloads == nil || len(build.Status.Workloads) == 0 {
		t.Fatal("No service builds running")
	}

	// ensure that build services are running

	for _, service := range build.Status.Workloads {
		if service.Status.State != v1.ContainerBuildStateRunning {
			t.Fatal("Service build should be in running state")
		}
	}

	fmt.Println("Service builds looking good!")

	// wait for build to finish
	fmt.Printf("Waiting for build %v to succeed\n", build.ID)

	waitFor(func() bool {
		build, err = latticeClient.Systems().Builds(mockSystemID).Get(build.ID)
		checkErr(err, t)
		return build.Status.State == v1.BuildStateSucceeded
	}, t)

	fmt.Printf("Build %v succeeded!\n", build.ID)

	fmt.Printf("Ensure that deploy %v enters in progress state!\n", deploy.ID)
	// ensure that deploy enters in progress state as well
	waitFor(func() bool {
		deploy, err = latticeClient.Systems().Deploys(mockSystemID).Get(deploy.ID)
		checkErr(err, t)
		return deploy.Status.State == v1.DeployStateInProgress
	}, t)

	// ensure that deploy state has succeeded
	fmt.Println("Wait until deploy succeeds")
	waitFor(func() bool {
		deploy, err = latticeClient.Systems().Deploys(mockSystemID).Get(deploy.ID)
		checkErr(err, t)
		return deploy.Status.State == v1.DeployStateSucceeded
	}, t)

	fmt.Println("Deploy succeeded!")

	// check service builds succeeded
	fmt.Println("Ensuring that service builds has succeeded...")
	for _, service := range build.Status.Workloads {
		if service.Status.State != v1.ContainerBuildStateSucceeded {
			t.Fatal("Service build should be in running state")
		}
	}
	fmt.Println("Service builds succeeded!")

	// list deploys
	deploys, err := latticeClient.Systems().Deploys(mockSystemID).List()
	checkErr(err, t)

	if len(deploys) != 1 {
		t.Fatal("bad # of elements for list deploys")
	}

	if deploys[0].ID != deploy.ID {
		t.Fatal("bad list deploy contents")
	}

	// test build logs
	fmt.Println("Test Build logs")
	reader, err := latticeClient.Systems().Builds(mockSystemID).Logs(build.ID, mockServicePath, nil, nil)
	checkErr(err, t)
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	if "this is a long line" != buf.String() {
		t.Fatal("Failed to get build logs")
	}

	// test service logs
	fmt.Println("Test Service logs")
	services, err := latticeClient.Systems().Services(mockSystemID).List()
	if err != nil {
		t.Fatalf("got err while retrieving system: %v", err)
	}

	if len(services) == 0 {
		t.Fatal("no services listed")
	}

	reader, err = latticeClient.Systems().Services(mockSystemID).Logs(
		services[0].ID,
		nil,
		nil,
		nil,
	)
	checkErr(err, t)
	buf = new(bytes.Buffer)
	buf.ReadFrom(reader)
	if "this is a long line" != buf.String() {
		t.Fatal("Failed to get service logs")
	}
}

func ensureSingleDeploy(t *testing.T) {
	fmt.Println("Ensure that system can have one accepted/running deploy at time")
	build, err := latticeClient.Systems().Builds(mockSystemID).CreateFromVersion(mockSystemVersion)
	checkErr(err, t)

	// Create first deploy
	fmt.Println("Creating first Depoly")
	deploy, err := latticeClient.Systems().Deploys(mockSystemID).CreateFromBuild(build.ID)
	checkErr(err, t)

	fmt.Printf("Created deploy %v\n", deploy.ID)

	// wait for build to run
	fmt.Printf("Waiting for deploy %v to enter accepted state\n", deploy.ID)
	waitFor(func() bool {
		deploy, err = latticeClient.Systems().Deploys(mockSystemID).Get(deploy.ID)
		checkErr(err, t)
		return deploy.Status.State == v1.DeployStateAccepted
	}, t)

	fmt.Printf("Deploy %v is in accepted state!\n", deploy.ID)
	fmt.Println("Attempt to create another deploy which should fail since there is one that is already accepted")
	deploy2, err := latticeClient.Systems().Deploys(mockSystemID).CreateFromBuild(build.ID)

	// wait for deploy to fail
	fmt.Printf("Waiting for deploy %v to enter failed state\n", deploy2.ID)
	waitFor(func() bool {
		deploy2, err = latticeClient.Systems().Deploys(mockSystemID).Get(deploy2.ID)
		checkErr(err, t)
		return deploy2.Status.State == v1.DeployStateFailed
	}, t)

	fmt.Printf("Deploy %v failed as expected!\n", deploy2.ID)

	fmt.Printf("Waiting for deploy %v to finish\n", deploy.ID)
	waitFor(func() bool {
		deploy, err = latticeClient.Systems().Deploys(mockSystemID).Get(deploy.ID)
		checkErr(err, t)
		return deploy.Status.State == v1.DeployStateSucceeded
	}, t)

}

func runJob(t *testing.T) {
	// create build
	fmt.Println("Test Run Job")
	cmd := []string{"echo", "foo"}
	env := definitionv1.ContainerEnvironment{}
	job, err := latticeClient.Systems().Jobs(mockSystemID).Create(mockServicePath, cmd, env)
	checkErr(err, t)

	fmt.Printf("Successfully created job. ID %v\n", job.ID)

	if job.Status.State != v1.JobStatePending {
		t.Fatalf("Job state is not pending")
	}

	// get job
	job1, err := latticeClient.Systems().Jobs(mockSystemID).Get(job.ID)
	checkErr(err, t)

	if job1 == nil {
		t.Fatal("Got job as nil")
	}

	// list jobs
	jobs, err := latticeClient.Systems().Jobs(mockSystemID).List()

	if len(jobs) != 1 {
		t.Fatal("bad # of elements for list jobs")
	}

	if jobs[0].ID != job.ID {
		t.Fatal("bad list jobs contents")
	}

	// wait for job to run
	fmt.Printf("Waiting for job %v to enter running state\n", job.ID)
	waitFor(func() bool {
		job, err = latticeClient.Systems().Jobs(mockSystemID).Get(job.ID)
		checkErr(err, t)
		return job.Status.State == v1.JobStateRunning
	}, t)

	fmt.Printf("job %v is in running state!\n", job.ID)

	// wait for job to finish
	fmt.Printf("Waiting for job %v to succeed\n", job.ID)

	waitFor(func() bool {
		job, err = latticeClient.Systems().Jobs(mockSystemID).Get(job.ID)
		checkErr(err, t)
		return job.Status.State == v1.JobStateSucceeded
	}, t)

	fmt.Printf("Job %v succeeded!\n", job.ID)

	// test job logs
	fmt.Println("Test Job logs")
	reader, err := latticeClient.Systems().Jobs(mockSystemID).Logs(job.ID, nil, nil)
	checkErr(err, t)
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	if "this is a long line" != buf.String() {
		t.Fatal("Failed to get job logs")
	}

}

func testSecrets(t *testing.T) {
	fmt.Println("Testing secrets...")
	path, _ := tree.NewPathSubcomponent("/test:x")
	fmt.Println("set secret")
	err := latticeClient.Systems().Secrets(mockSystemID).Set(path, "1")
	checkErr(err, t)
	secrets, err := latticeClient.Systems().Secrets(mockSystemID).List()
	checkErr(err, t)

	fmt.Println("list secrets")
	if secrets == nil || len(secrets) != 1 {
		t.Fatal("Wrong number of secrets.")
	}

	fmt.Println("get secret")
	secret, err := latticeClient.Systems().Secrets(mockSystemID).Get(path)
	checkErr(err, t)
	if secret.Value != "1" {
		t.Fatal("Bad secret.")
	}

	fmt.Println("unset secret")
	err = latticeClient.Systems().Secrets(mockSystemID).Unset(path)
	checkErr(err, t)

	secrets, err = latticeClient.Systems().Secrets(mockSystemID).List()
	checkErr(err, t)

	if len(secrets) != 0 {
		t.Fatal("Secret was not unset")
	}
}

func checkSystemHealth(t *testing.T) {
	// ensure that system services are up
	services, err := latticeClient.Systems().Services(mockSystemID).List()
	checkErr(err, t)
	if len(services) == 0 {
		t.Fatalf("no services in the system")
	}

	for _, service := range services {
		if service.Status.State != v1.ServiceStateStable {
			t.Fatalf("Service state is not stable")
		}
	}
}

func teardownSystem(t *testing.T) {

	fmt.Println("Tearing system down...")

	teardown, err := latticeClient.Systems().Teardowns(mockSystemID).Create()
	checkErr(err, t)

	fmt.Printf("Created teardown %v", teardown.ID)

	if teardown.Status.State != v1.TeardownStatePending {
		t.Fatalf("teardown state is not pending")
	}

	// wait for teardown run
	fmt.Printf("Waiting for teardown %v to run\n", teardown.ID)

	waitFor(func() bool {
		teardown, err = latticeClient.Systems().Teardowns(mockSystemID).Get(teardown.ID)
		checkErr(err, t)
		return teardown.Status.State == v1.TeardownStateInProgress
	}, t)

	fmt.Printf("Teardown %v entered the in progress state!\n", teardown.ID)

	// wait for teardown succeed
	fmt.Printf("Waiting for teardown %v to succeed\n", teardown.ID)

	waitFor(func() bool {
		teardown, err = latticeClient.Systems().Teardowns(mockSystemID).Get(teardown.ID)
		checkErr(err, t)
		return teardown.Status.State == v1.TeardownStateSucceeded
	}, t)

	// check that system services are nil after teardown
	fmt.Println("Checking that system services are down after teardown...")
	services, err := latticeClient.Systems().Services(mockSystemID).List()
	if len(services) != 0 {
		t.Fatal("System services still up")
	}

	fmt.Printf("Teardown %v succeeded!\n", teardown.ID)
}

func deleteSystem(t *testing.T) {
	system, err := latticeClient.Systems().Get(mockSystemID)
	checkErr(err, t)

	if system == nil {
		t.Fatal("System not found")
	}

	fmt.Println("Deleting system...")
	latticeClient.Systems().Delete(mockSystemID)

	_, err = latticeClient.Systems().Get(mockSystemID)
	v1err, ok := err.(*v1.Error)
	if !ok {
		t.Fatal("Expected an invalid system error")
	}

	if v1err.Code != v1.ErrorCodeInvalidSystemID {
		t.Fatal("Expected an invalid system error")
	}

	fmt.Println("System deleted!")
}

func testInvalidIDs(t *testing.T) {
	fmt.Println("Test invalid IDs")

	// test invalid system
	{
		fmt.Println("Test invalid system")
		_, err := latticeClient.Systems().Get("no-such-system")
		v1err, ok := err.(*v1.Error)
		if !ok {
			t.Fatal("Expected an invalid system error")
		}

		if v1err.Code != v1.ErrorCodeInvalidSystemID {
			t.Fatal("Expected an invalid system error")
		}
	}

	// test other stuff
	testID := v1.SystemID("test")
	_, err := latticeClient.Systems().Create(testID, mockSystemDefURL)
	checkErr(err, t)
	waitFor(func() bool {
		s, err := latticeClient.Systems().Get(testID)
		checkErr(err, t)
		return s.Status.State == v1.SystemStateStable
	}, t)

	// test invalid build id error
	{
		_, err = latticeClient.Systems().Builds(testID).Get("bad-build")
		v1err, ok := err.(*v1.Error)
		if !ok {
			t.Fatal("Expected an invalid build error")
		}

		if v1err.Code != v1.ErrorCodeInvalidBuildID {
			t.Fatal("Expected an invalid build error")
		}
	}

	// test invalid deploy id error
	{
		_, err = latticeClient.Systems().Deploys(testID).Get("bad-deploy")
		v1err, ok := err.(*v1.Error)
		if !ok {
			t.Fatal("Expected an invalid deploy error")
		}

		if v1err.Code != v1.ErrorCodeInvalidDeployID {
			t.Fatal("Expected an invalid deploy error")
		}
	}

	// test invalid teardown id error
	{
		_, err = latticeClient.Systems().Teardowns(testID).Get("bad-teardown")
		v1err, ok := err.(*v1.Error)
		if !ok {
			t.Fatal("Expected an invalid teardown error")
		}

		if v1err.Code != v1.ErrorCodeInvalidTeardownID {
			t.Fatal("Expected an invalid teardown error")
		}
	}

	latticeClient.Systems().Delete("test")
}

func testInvalidDefinition(t *testing.T) {
	fmt.Println("Test invalid definition URL")

	testID := v1.SystemID("test")
	_, err := latticeClient.Systems().Create("test", "xxxxxxx")
	checkErr(err, t)

	waitFor(func() bool {
		s, err := latticeClient.Systems().Get(testID)
		checkErr(err, t)
		return s.Status.State == v1.SystemStateStable
	}, t)

	build, err := latticeClient.Systems().Builds(testID).CreateFromVersion(mockSystemVersion)
	checkErr(err, t)

	waitFor(func() bool {
		b, err := latticeClient.Systems().Builds(testID).Get(build.ID)
		checkErr(err, t)
		return b.Status.State == v1.BuildStateFailed
	}, t)

	fmt.Println("invalid build failed as expected")
}

func authTest(t *testing.T) {
	fmt.Println("Test authentication")

	fmt.Println("Testing auth with good API key")
	_, err := latticeClient.Systems().List()

	if err != nil {
		t.Fatal("Failed to authenticate")
	}

	fmt.Println("Auth success!!")

	fmt.Println("Testing auth with bad API key")
	badClient := clientrest.NewBearerTokenClient(mockAPIServerURL, "bad api key").V1()
	_, err = badClient.Systems().List()

	if err != nil && !strings.Contains(fmt.Sprintf("%v", err), "status code 403") {
		t.Fatal("Expected an authentication error")
	}
	fmt.Printf("Got an expected authentication failure error: %v\n", err)

}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
}

func setupMockTest() {
	fmt.Println("Setting up test. Starting API Server")

	os.RemoveAll(mockRepoPath)

	err := git.Init(mockRepoPath)
	if err != nil {
		panic(err)
	}

	hash, err := git.WriteAndCommitFile(mockRepoPath, resolver.DefaultFile, mockSystemBytes, 0700, "commit")
	if err != nil {
		panic(err)
	}

	err = git.Tag(mockRepoPath, hash, string(mockSystemVersion))
	if err != nil {
		panic(err)
	}

	// run api server
	gitResolver, err := git.NewResolver("/tmp/lattice/test/api/server/rest/resolver", true)
	if err != nil {
		panic(err)
	}

	r := resolver.NewComponentResolver(
		gitResolver,
		mockresolver.NewMemoryTemplateStore(),
		mockresolver.NewMemorySecretStore(),
	)

	b := mockbackend.NewMockBackend(r)
	go RunNewRestServer(b, r, mockServerAPIPort, mockServerAPIKey)
	fmt.Println("API server started")

	// wait a second to let server start
	time.Sleep(time.Second)
}

func waitFor(condition func() bool, t *testing.T) {
	for i := 0; i <= 200; i++ {
		if !condition() {
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}

	if !condition() {
		t.Fatal("Timed out")
	}
}
