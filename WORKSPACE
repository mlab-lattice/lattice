load("//bazel/workspace:bazel_dependencies.bzl", "lattice_bazel_dependencies")

lattice_bazel_dependencies()

# Must load our go dependencies before pulling in rules_go's dependencies so that
# we can specify the proper versions of repositories that are used by go_rules_dependencies:
# https://github.com/bazelbuild/rules_go/blob/0.9.0/go/workspace.rst#go_rules_dependencies
load("//bazel/workspace:dependencies.bzl", "lattice_dependencies")

lattice_dependencies()

load("//bazel/workspace:initialize.bzl", "initialize_lattice_bazel_dependencies")

initialize_lattice_bazel_dependencies()
