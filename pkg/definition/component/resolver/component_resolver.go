package resolver

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const (
	DepthInfinite = -1

	fileExtensionJSON = ".json"
	fileExtensionYAML = ".yaml"
	fileExtensionYML  = ".yml"

	DefaultFile = "lattice.yaml"
)

type Interface interface {
	Resolve(
		c component.Interface,
		id v1.SystemID,
		path tree.Path,
		ctx *git.CommitReference,
		depth int,
	) (*ResolutionTree, error)
	Versions(component.Interface, semver.Range) ([]string, error)
}

func NewComponentResolver(
	gitResolver *git.Resolver,
	templateStore TemplateStore,
	secretStore SecretStore,
) Interface {
	r := &componentResolver{
		gitResolver: gitResolver,
	}

	v1Resolver := newV1ComponentResolver(r, gitResolver, templateStore, secretStore)
	r.v1 = v1Resolver
	return r
}

type componentResolver struct {
	gitResolver *git.Resolver
	v1          *v1ComponentResolver
}

func (r *componentResolver) Resolve(
	c component.Interface,
	id v1.SystemID,
	path tree.Path,
	ctx *git.CommitReference,
	depth int,
) (*ResolutionTree, error) {
	// TODO(kevindrosendahl): this here is why private system definitions aren't supported
	rctx := &resolutionContext{CommitReference: ctx}
	result := NewResolutionTree()
	err := r.resolve(c, id, path, rctx, depth, result)
	return result, err
}

func (r *componentResolver) Versions(c component.Interface, rng semver.Range) ([]string, error) {
	switch typed := c.(type) {
	case *definitionv1.Reference:
		return r.v1.Versions(typed, rng)

	default:
		return nil, fmt.Errorf("cannot list versions of type %v", c.Type().String())
	}
}

func (r *componentResolver) resolve(
	c component.Interface,
	id v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	depth int,
	result *ResolutionTree,
) error {
	switch c.Type().APIVersion {
	case definitionv1.APIVersion:
		return r.v1.Resolve(c, id, path, ctx, depth, result)

	default:
		return fmt.Errorf("unsupported component type: %v", c.Type().String())
	}
}

func (r *componentResolver) newComponent(m map[string]interface{}) (component.Interface, error) {
	t, err := component.TypeFromMap(m)
	if err != nil {
		return nil, err
	}

	switch t.APIVersion {
	case definitionv1.APIVersion:
		return definitionv1.NewComponent(m)

	default:
		return nil, fmt.Errorf("invalid type api %v", t.APIVersion)
	}
}
