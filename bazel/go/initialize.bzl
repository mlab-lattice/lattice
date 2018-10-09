load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
def initialize_rules_go():
  go_rules_dependencies()
  go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
def initialize_bazel_gazelle():
  gazelle_dependencies()
