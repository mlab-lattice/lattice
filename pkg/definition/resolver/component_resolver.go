package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type ComponentResolver interface {
	ResolveComponent(
		path tree.NodePath,
		ctx *definitionv1.GitRepositoryReference,
		ref *definitionv1.Reference,
	) (component.Interface, error)
}

type DefaultComponentResolver struct {
	referenceResolver ReferenceResolver
}

func NewComponentResolver(referenceResolver ReferenceResolver) ComponentResolver {
	return &DefaultComponentResolver{referenceResolver}
}

func (r *DefaultComponentResolver) ResolveComponent(
	path tree.NodePath,
	ctx *definitionv1.GitRepositoryReference,
	ref *definitionv1.Reference,
) (component.Interface, error) {
	c, resolvedContext, err := r.referenceResolver.ResolveReference(path, ctx, ref)
	if err != nil {
		return nil, err
	}

	// If the reference resolved to another reference, resolve that reference.
	// FIXME(kevinrosendahl): detect cycles
	if ref2, ok := c.(*definitionv1.Reference); ok {
		return r.ResolveComponent(path, resolvedContext, ref2)
	}

	// If the reference resolved to a system, resolve the system's components.
	if system, ok := c.(*definitionv1.System); ok {
		return r.resolveSystemComponents(path, resolvedContext, system)
	}

	// Otherwise the reference resolved to a leaf, return the component.
	return c, nil
}

func (r *DefaultComponentResolver) resolveSystemComponents(
	path tree.NodePath,
	ctx *definitionv1.GitRepositoryReference,
	system *definitionv1.System,
) (*definitionv1.System, error) {
	// Loop through each of the components.
	//  - If the component is a system, recursively resolve its components.
	//  - If the component is a reference, resolve it (potentially also recursively resolving
	//    system components if the reference was to a system).
	for name, c := range system.Components {
		childPath := path.Child(name)
		switch typedComponent := c.(type) {

		case *definitionv1.System:
			// If the component is a system, recursively resolve the system and overwrite it in the components map
			subSystem, err := r.resolveSystemComponents(childPath, ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			system.Components[name] = subSystem

		case *definitionv1.Reference:
			// If the component is a reference, resolve the reference.
			resolved, err := r.ResolveComponent(childPath, ctx, typedComponent)
			if err != nil {
				return nil, err
			}

			system.Components[name] = resolved
		}
	}

	return system, nil
}
