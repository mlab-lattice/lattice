# Repository Layout

## bazel
The `bazel` directory contains rule definitions for [`Bazel`](https://bazel.build). It may be helpful to read the building [documentation](building.md).

## ci
The `ci` directory contains configuration and scripts for lattice CI pipelines.

## cmd
The `cmd` directory contains `go` code for lattice binaries.

In general, these `go` packages should as thin as possible, mainly parsing arguments and flags and delegating to libraries.

## docker
The `docker` directory contains `Bazel` targets for building, pushing, and running docker images. It may be helpful to read the docker images [documentation](docker-images.md).

## docs
The `docs` directory contains documentation.

## hack
The `hack` directory contains scripts mostly for development.

## pkg
The `pkg` directory contains all `go` libraries. For a more in depth exploration of the libraries, read the architecture [documentation](architecture).

One important general pattern of the lattice packages is that usually when an interface is defined, the implementation of the interface is housed under a path reflecting the path of the interface.

For example, `pkg/api/server/backend.Interface` is implemented in `pkg/backend/kubernetes/api/server/backend` and `pkg/backend/mock/api/server/backend`.

## terraform
The `terraform` directory contains `terrafrom` modules used by lattice. This may soon be removed.
