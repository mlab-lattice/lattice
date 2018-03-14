load(":bazel/repositories.bzl", "rules_go_dependencies", "bazel_gazelle_dependencies", "rules_docker_dependencies", "rules_package_manager_dependencies")
rules_go_dependencies()
bazel_gazelle_dependencies()
rules_docker_dependencies()
rules_package_manager_dependencies()

# Must load our go dependencies before pulling in rules_go's dependencies so that
# we can specify the proper versions of repositories that are used by go_rules_dependencies:
# https://github.com/bazelbuild/rules_go/blob/0.9.0/go/workspace.rst#go_rules_dependencies
load(":bazel/dependencies.bzl", "go_dependencies", "docker_dependencies")
go_dependencies()
docker_dependencies()

load(":bazel/initialize.bzl", "initialize_rules_go", "initialize_bazel_gazelle", "initialize_rules_docker", "initialize_rules_package_manager")
initialize_rules_go()
initialize_bazel_gazelle()
initialize_rules_docker()
initialize_rules_package_manager()
