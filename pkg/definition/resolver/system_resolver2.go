package resolver

import (
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type SystemResolver2 interface {
	ResolveSystem(ctx *definitionv1.GitRepositoryReference, system *definitionv1.System) (*definitionv1.System, error)
}

type DefaultSystemResolver struct {
	referenceResolver ReferenceResolver
}

func NewSystemResolver2(referenceResolver ReferenceResolver) SystemResolver2 {
	return &DefaultSystemResolver{referenceResolver}
}

func (r *DefaultSystemResolver) ResolveSystem(
	ctx *definitionv1.GitRepositoryReference,
	system *definitionv1.System,
) (*definitionv1.System, error) {
	s := &definitionv1.System{}
	*s = *system

	for name, c := range system.Components {
		switch typedComponent := c.(type) {

		// If the component is a system, recursively resolve the system and overwrite it in the components map
		case *definitionv1.System:
			subSystem, err := r.ResolveSystem(ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			s.Components[name] = subSystem

		// If the component is a reference, resolve the reference. If the reference ended up being to a system,
		// recursively resolve the system as well.
		case *definitionv1.Reference:
			resolved, err := r.referenceResolver.ResolveReference(ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			// If it did not resolve to a system, no more work to do.
			resolvedSys, ok := resolved.(*definitionv1.System)
			if !ok {
				s.Components[name] = resolved
				break
			}

			// If it did resolve to a system, recursively resolve the system.
			// First generate the proper context for the system to be resolved in.
			var rCtx *definitionv1.GitRepositoryReference
			switch {

			// The reference was resolved from the same git repository as the
			// original system of this function context.
			case typedComponent.File != nil:
				rCtx = &definitionv1.GitRepositoryReference{
					GitRepository: ctx.GitRepository,
					File:          *typedComponent.File,
				}

			// The reference was resolved from a new git repository, so the
			// system should be resolved in the context of the new git repository.
			case typedComponent.GitRepository != nil:
				rCtx = &definitionv1.GitRepositoryReference{}
				*rCtx = *typedComponent.GitRepository
			}

			subSystem, err := r.ResolveSystem(ctx, resolvedSys)
			if err != nil {
				return nil, err
			}

			s.Components[name] = subSystem
		}
	}

	return s, nil
}
