def rules_go_dependencies():
  native.git_repository(
      name = "io_bazel_rules_go",
      remote = "https://github.com/bazelbuild/rules_go.git",
      commit = "bfb156af43d2eabdabfcd2a9b668b2d6bf708fbb",
  )

def bazel_gazelle_dependencies():
  native.http_archive(
      name = "bazel_gazelle",
      url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.8/bazel-gazelle-0.8.tar.gz",
      sha256 = "e3dadf036c769d1f40603b86ae1f0f90d11837116022d9b06e4cd88cae786676",
  )

def rules_docker_dependencies():
  native.git_repository(
      name = "io_bazel_rules_docker",
      remote = "https://github.com/bazelbuild/rules_docker.git",
      commit = "3caf72f166f8b6b0e529442477a74871ad4d35e9",
  )

def rules_package_manager_dependencies():
  native.git_repository(
      name = "distroless",
      remote = "https://github.com/GoogleCloudPlatform/distroless.git",
      commit = "e5854b38a12bb37adaf0edb193f97b32a3bcaee0",
  )
