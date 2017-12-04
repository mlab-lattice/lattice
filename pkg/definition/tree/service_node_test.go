package tree

import (
	"testing"

	definitiontest "github.com/mlab-lattice/system/pkg/definition/test"
	testutil "github.com/mlab-lattice/system/pkg/util/test"
)

func TestServiceNode(t *testing.T) {
	sd := definitiontest.MockService()

	s, err := NewServiceNode(sd, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	testutil.ValidateToJSON(t, "ServiceNode", s, definitiontest.MockServiceExpectedJSON())

	if s.Parent() != nil {
		t.Error("Parent() != nil")
	}

	if s.Definition() != s.definition {
		t.Error("Interface() != s.ServiceDefinition")
	}

	if len(s.Subsystems()) != 0 {
		t.Errorf("Subsystems() not empty")
	}
}
