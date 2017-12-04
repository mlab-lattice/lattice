load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

def initialize_rules_go():
  go_rules_dependencies()
  go_register_toolchains()
  proto_register_toolchains()

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")

def initialize_rules_docker():
  _go_image_repos()


load("@distroless//package_manager:package_manager.bzl", "package_manager_repositories",)

def initialize_rules_package_manager():
  package_manager_repositories()
