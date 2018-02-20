def rules_go_dependencies():
  native.git_repository(
     name = "io_bazel_rules_go",
     remote = "https://github.com/bazelbuild/rules_go.git",
     commit = "ff7e3364d9383cf14155f8c2efc87218d07eb03b",
  )

def bazel_gazelle_dependencies():
  native.http_archive(
      name = "bazel_gazelle",
      url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.9/bazel-gazelle-0.9.tar.gz",
      sha256 = "0103991d994db55b3b5d7b06336f8ae355739635e0c2379dea16b8213ea5a223",
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
