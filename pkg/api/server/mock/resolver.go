package mock

import (
	"github.com/blang/semver"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type DefaultMockComponentResolver struct {
}

func newMockComponentResolver() resolver.ComponentResolver {
	return &DefaultMockComponentResolver{}
}

func (*DefaultMockComponentResolver) ResolveReference(
	systemID v1.SystemID,
	path tree.Path,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	depth int32,
) (*resolver.ResolutionResult, error) {

	return &resolver.ResolutionResult{
		Component: getMockSystem(),
	}, nil
}

func (*DefaultMockComponentResolver) Versions(repository string, semverRange semver.Range) ([]string, error) {
	return []string{"1.0.0", "2.0.0"}, nil
}

func getMockSystem() *definitionv1.System {

	return &definitionv1.System{
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

}
