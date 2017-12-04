def rules_go_dependencies():
  native.git_repository(
      name = "io_bazel_rules_go",
      remote = "https://github.com/bazelbuild/rules_go.git",
      commit = "44b3bdf7d3645cbf0cfd786c5f105d0af4cf49ca",
  )

def rules_docker_dependencies():
  native.git_repository(
      name = "io_bazel_rules_docker",
      remote = "https://github.com/bazelbuild/rules_docker.git",
      commit = "caa55f1e00e5909dbd689f298e2c6d3ef3e65d81",
  )

def rules_package_manager_dependencies():
  native.git_repository(
      name = "distroless",
      remote = "https://github.com/GoogleCloudPlatform/distroless.git",
      commit = "e5854b38a12bb37adaf0edb193f97b32a3bcaee0",
  )
