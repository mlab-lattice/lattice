package mock

import (
	"fmt"
	"testing"

	latticerest "github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

const (
	mockSystemId     = v1.SystemID("mock-system")
	mockSystemDefURL = "https://github.com/mlab-lattice/mock-system.git"
	mockAPIServerURL = "http://localhost:8876"
)

func TestMockServer(t *testing.T) {
	setupMockTest()
	t.Run("TestMockServer", mockTests)
}

func mockTests(t *testing.T) {
	latticeClient := latticerest.NewClient(mockAPIServerURL).V1()

	// test create system
	fmt.Println("Test create system")
	system, err := latticeClient.Systems().Create(mockSystemId, mockSystemDefURL)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
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
	s, err := latticeClient.Systems().Get(system.ID)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if s.ID != mockSystemId {
		t.Fatalf("Bad system returned from Get")
	}

	// test list system
	fmt.Println("Test List Systems")
	systems, err := latticeClient.Systems().List()
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if len(systems) != 1 {
		t.Fatalf("Wrong number of systems")
	}

	if systems[0].ID != mockSystemId {
		t.Fatalf("bad list systems")
	}

}

func setupMockTest() {
	fmt.Println("Setting up test. Starting API Server")
	// run api server
	go RunMockNewRestServer()

	fmt.Println("API server started")
}
