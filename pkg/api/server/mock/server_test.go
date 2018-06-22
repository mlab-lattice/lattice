package mock

import (
	"fmt"
	"testing"

	latticerest "github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func TestMockServer(t *testing.T) {
	setupMockTest()
	t.Run("TestMockServer", mockTests)
}

func mockTests(t *testing.T) {
	latticeClient := latticerest.NewClient("http://localhost:8876").V1()

	fmt.Println("Test create system")
	system, err := latticeClient.Systems().Create(v1.SystemID("mock-system"),
		"https://github.com/mlab-lattice/mock-system.git")

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if system.ID != v1.SystemID("mock-system") {
		t.Fatalf("Bad system returned from create")
	}
	fmt.Println("System created successfully")

	fmt.Println("Test Get System")
	s, err := latticeClient.Systems().Get(system.ID)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if s.ID != v1.SystemID("mock-system") {
		t.Fatalf("Bad system returned from Get")
	}

}

func setupMockTest() {
	fmt.Println("Setting up test. Starting API Server")
	// run api server
	go RunMockNewRestServer()

	fmt.Println("API server started")
}
