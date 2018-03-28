def rules_go_dependencies():
  native.http_archive(
      name = "io_bazel_rules_go",
      url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
      sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
  )

def bazel_gazelle_dependencies():
  native.http_archive(
      name = "bazel_gazelle",
      url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.1/bazel-gazelle-0.10.1.tar.gz",
      sha256 = "d03625db67e9fb0905bbd206fa97e32ae9da894fe234a493e7517fd25faec914",
  )

def rules_docker_dependencies():
  native.git_repository(
      name = "io_bazel_rules_docker",
      remote = "https://github.com/bazelbuild/rules_docker.git",
      commit = "c7f9eaa63bc3a31acab5e399c72b4e5228ab5ad7",
  )

def rules_package_manager_dependencies():
  native.git_repository(
      name = "distroless",
      remote = "https://github.com/GoogleCloudPlatform/distroless.git",
      commit = "e5854b38a12bb37adaf0edb193f97b32a3bcaee0",
  )
