package mock

import (
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type MockSystemResolver struct {
}

func newMockSystemResolver() *MockSystemResolver {
	return &MockSystemResolver{}
}

func (*MockSystemResolver) ResolveDefinition(uri string, gitResolveOptions *git.Options) (tree.Node, error) {
	return getMockSystemDefinition()
}

func (*MockSystemResolver) ListDefinitionVersions(uri string, gitResolveOptions *git.Options) ([]string, error) {
	return []string{"1.0.0", "2.0.0"}, nil
}

func getMockSystemDefinition() (tree.Node, error) {

	system := &definitionv1.System{
		Description: "Mock System",
		Components: map[string]component.Interface{
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
		},
	}

	return definitionv1.NewNode(system, "mock-system", nil)

}
