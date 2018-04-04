load("//bazel/docker:initialize.bzl", "initialize_rules_docker", "initialize_rules_package_manager")
load("//bazel/go:initialize.bzl", "initialize_rules_go", "initialize_bazel_gazelle")

def initialize_lattice_bazel_dependencies():
  initialize_rules_go()
  initialize_bazel_gazelle()
  initialize_rules_docker()
  initialize_rules_package_manager()
