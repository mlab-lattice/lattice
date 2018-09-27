load("@io_bazel_rules_docker//go:image.bzl", go_image_repositories="repositories")
load("@io_bazel_rules_docker//container:container.bzl", container_repositories = "repositories")
def initialize_rules_docker():
  container_repositories()
  go_image_repositories()

load("@distroless//package_manager:package_manager.bzl", "package_manager_repositories",)
def initialize_rules_package_manager():
  package_manager_repositories()
