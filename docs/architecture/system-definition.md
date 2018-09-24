# System Definition

This document attempts to outline the components of the lattice code base used for working with system definitions. It will assume a working knowledge of the user-facing details of the definition language.

## `tree`

At a high level, a `system` defines a tree of components. As such, lattice provides a number of utilities for working with trees in the [`pkg/definition/tree`](../../pkg/definition/tree) package.

More information about the comonents discussed below can be found in the package's [documentation](https://godoc.org/github.com/mlab-lattice/lattice/pkg/definition/tree).

### `tree.Path`

A `tree.Path` is a `/slash/seperated/path` that is used in lattice to denote a component's location in a system tree.

### `tree.Subcomponent`

A `tree.Subcomponent` is a `/tree/path:with-a-selector`. This can be thought of a namespace scoped to the given path. The `/tree/path` is referred to as the `tree.Subcomponent`'s `Path`, and the `with-a-selector` is referred to as the `Subcomponent`.

This is generally used for things such as `secrets` (e.g. `/my/service:mongo-url`).

### `tree.Radix`

A `tree.Radix` is an efficient data structure for associating information with `tree.Path`s, as well as querying and manipulating that information.

### `tree.JSONRadix`

A `tree.JSONRadix` is a `tree.Radix` with additional information allowing it to (de)serialize to/from JSON.

## `component`

The [`component.Interface`](../../pkg/definition/component/interface.go) essentially establishes what it means to be a system component: having a `type` field with a value matching `api-version/type-name`.

As of writing, only well known types actually are able to resolve into being a `component.Interface`.

### Versioning

As mentioned above, a `component.Interface` is something that has a `type`. Currently there is only one valid API version recognized in lattice: `v1`.

#### v1

The structs defining valid `v1` `components` are defined in [`pkg/definition/v1`](../../pkg/definition/v1). This package is commonly referred to as `definitionv1` to avoid confusion with the [`pkg/api/v1`](../../pkg/api/v1) package, which is commonly referred to as `v1`.

Importantly, these structs define the valid structures of fully _resolved_ `components` (more on this below).

## `component/resolver`

### `resolver.Interface`

`resolver.Interface` defines an interface that "resolves" `component.Interface`s.

Currently, that essentially means:
- if the component is a `v1/reference`
  - attempt to retrieve the `template` from wherever it is defined to be located
  - inject any `parameters` into the `template` (the package that deals with this is located at [`pkg/definition/component/resolver/template`](../../pkg/definition/component/resolver/template))
  - attempt to create a `component.Interface` out of the `template` given the value of its `type` field
  - recursively call `Resolve` on the resolved `component.Interface`
- if the component is a `v1/system`
  - loop through all of its `components`, recursively calling `Resolve` on them
- else
  - noop
  
Note that when `Resolve` is called, the `resolver.Interface` dispatches the work based on the `component.Interface`'s `Type`'s `APIVersion`. And as each `v1/reference` and `v1/system` recurses back to the top level `Resolve` method, this in theory allows interop between different `APIVersion`s. So if we were to one day introduce a `definitionv2`, in theory a `v1/reference` could point at a `v2` `component`, or a `v1/system` could include `v2` `component`s, or a `v2` `component` could include a `v1` component, etc.

`resolver.Interface` also includes a `Versions` method. This currently returns an empty slice for anything except a `v1/reference` with a `git_repository`. For a `v1/reference` it will return all tags from the `git_repository` that are valid [`semver`](https://semver.org) versions.

### `resolver.ResolutionTree`

When called, the `resolver.Interface`'s `Resolve` method returns a [`resolver.ResolutionTree`](../../pkg/definition/component/resolver/resolution_tree.go).

The `resolver.ResolutionTree` is a `tree.JSONRadix` that contains [`resolver.ResolutionInfo`](../../pkg/definition/component/resolver/resolution_info.go) about each `component` that was resolved during the `Resolve` call.

The `resolver.ResolutionInfo` contains both the `component.Interface` resolved at that `tree.Path`, as well as information about how the `component` was resolved (e.g. information about the git repository the template lived in).

Note that the `component.Interface` in the `resolver.ResolutionInfo` contains `components` with `template`s fully resolved, but `v1/references` not resolved.

For example, if you had a system at the root (`/`) with two components `s` and `r`, where `s` is a `v1/service` and `r` is a `v1/reference` that points at a `v1/job`, the `resolver.ResolutionTree` would look logically like:

```
/:
  component:
    type: v1/system
    components:
      r:
        type: v1/reference
        ...
      s:
        type: v1/service
        ...
  ...
/r:
  component:
    type: v1/job
    ...
  ...
/s:
  component:
    type: v1/service
    ...
  ... 
```

Another thing to note is that `Resolve` takes a `tree.Path` argument, which specifies the `component.Interface` that is being resolved's location in the `system` tree. In the example above `/` was passed in, but if say `/foo/bar` had been passed in, the `resolver.ResolutionTree` returned would have logically looked like:

```
/foo/bar:
  component:
    type: v1/system
    components:
      r:
        type: v1/reference
        ...
      s:
        type: v1/service
        ...
  ...
/foo/bar/r:
  component:
    type: v1/job
    ...
  ...
/foo/bar/s:
  component:
    type: v1/service
    ...
  ... 
```

This allows consumers of the library to only `Resolve` a portion of an existing definition, and can use `tree.Radix` methods such as `ReplacePrefix` to replace a subtree in the existing definition.

This is in fact what is in fact what is occurring when a user calls

```
latticectl deploy --path /foo/bar
```

For more information on this please see the documentation on [Lifecycle Action Locking](lifecycle/system.md#lifecycle-action-locking).
