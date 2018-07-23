package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type SystemResolver2 interface {
	ResolveSystem(path tree.NodePath, ctx *definitionv1.GitRepositoryReference, system *definitionv1.System) (*definitionv1.System, error)
}

type DefaultSystemResolver struct {
	referenceResolver ReferenceResolver
}

func NewSystemResolver2(referenceResolver ReferenceResolver) SystemResolver2 {
	return &DefaultSystemResolver{referenceResolver}
}

func (r *DefaultSystemResolver) ResolveSystem(
	path tree.NodePath,
	ctx *definitionv1.GitRepositoryReference,
	system *definitionv1.System,
) (*definitionv1.System, error) {
	s := &definitionv1.System{}
	*s = *system

	for name, c := range system.Components {
		childPath := path.Child(name)
		switch typedComponent := c.(type) {

		// If the component is a system, recursively resolve the system and overwrite it in the components map
		case *definitionv1.System:
			subSystem, err := r.ResolveSystem(childPath, ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			s.Components[name] = subSystem

		// If the component is a reference, resolve the reference. If the reference ended up being to a system,
		// recursively resolve the system as well.
		case *definitionv1.Reference:
			resolved, resolvedCtx, err := r.referenceResolver.ResolveReference(childPath, ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			// If it did not resolve to a system, no more work to do.
			resolvedSys, ok := resolved.(*definitionv1.System)
			if !ok {
				s.Components[name] = resolved
				break
			}

			subSystem, err := r.ResolveSystem(childPath, resolvedCtx, resolvedSys)
			if err != nil {
				return nil, err
			}

			s.Components[name] = subSystem
		}
	}

	return s, nil
}
