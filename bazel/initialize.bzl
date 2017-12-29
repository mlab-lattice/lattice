load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
def initialize_rules_go():
  go_rules_dependencies()
  go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
def initialize_bazel_gazelle():
  gazelle_dependencies()

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")
def initialize_rules_docker():
  _go_image_repos()

load("@distroless//package_manager:package_manager.bzl", "package_manager_repositories",)
def initialize_rules_package_manager():
  package_manager_repositories()
