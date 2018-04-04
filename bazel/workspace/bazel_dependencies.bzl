load("//bazel/docker:bazel_dependencies.bzl", "rules_docker_dependencies", "rules_package_manager_dependencies")
load("//bazel/go:bazel_dependencies.bzl", "rules_go_dependencies", "bazel_gazelle_dependencies")

def lattice_bazel_dependencies():
  rules_go_dependencies()
  bazel_gazelle_dependencies()
  rules_docker_dependencies()
  rules_package_manager_dependencies()
