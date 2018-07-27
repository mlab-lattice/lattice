def rules_docker_dependencies():
  native.git_repository(
      name = "io_bazel_rules_docker",
      remote = "https://github.com/bazelbuild/rules_docker.git",
      commit = "bb2748d7258f0c15eada2382eab9674ab65f0521",
  )

def rules_package_manager_dependencies():
  native.git_repository(
      name = "distroless",
      remote = "https://github.com/GoogleCloudPlatform/distroless.git",
      commit = "3cda1707e86ac6160444fded894e712f85619c05",
  )
