load("//bazel/docker:dependencies.bzl", "docker_dependencies")
load("//bazel/go:dependencies.bzl", "go_dependencies")

def lattice_dependencies():
  go_dependencies()
  docker_dependencies()
