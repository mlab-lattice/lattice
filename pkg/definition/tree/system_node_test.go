package tree

import (
	"testing"

	definitiontest "github.com/mlab-lattice/system/pkg/definition/test"
	testutil "github.com/mlab-lattice/system/pkg/util/test"
)

func TestSystemNode(t *testing.T) {
	sd := definitiontest.MockSystem()

	s, err := NewSystemNode(sd, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	testutil.ValidateToJSON(t, "SystemNode", s, definitiontest.MockSystemExpectedJSON())

	if s.Parent() != nil {
		t.Error("Parent() != nil")
	}

	if s.Definition() != s.definition {
		t.Error("Interface() != s.ServiceDefinition")
	}

	if len(s.Subsystems()) != 1 {
		t.Errorf("Subsystems() does not have 1 element")
	}

	child, ok := s.Subsystems()["/"+NodePath(s.definition.Meta.Name+"/"+s.definition.Subsystems[0].Metadata().Name)]
	if !ok {
		t.Fatal("Subsystems() does not contain child service")
	}

	if child.Parent() != Node(s) {
		t.Errorf("child service Parent() != Node(system)")
	}
}
