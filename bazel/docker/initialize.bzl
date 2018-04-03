load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos="repositories")
def initialize_rules_docker():
  _go_image_repos()

load("@distroless//package_manager:package_manager.bzl", "package_manager_repositories",)
def initialize_rules_package_manager():
  package_manager_repositories()
