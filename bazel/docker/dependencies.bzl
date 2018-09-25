load("@distroless//package_manager:package_manager.bzl", "dpkg_src", "dpkg_list")
load("@io_bazel_rules_docker//container:container.bzl", "container_pull")

def docker_dependencies():
  _docker_dependencies_debian_pkg()
  _docker_dependencies_debian_slim_image()
  _docker_dependencies_helm()
  _docker_dependencies_kubectl()
  _docker_dependencies_terraform()

def _docker_dependencies_debian_pkg():
  dpkg_src(
      name = "debian_stretch",
      arch = "amd64",
      distro = "stretch",
      sha256 = "9aea0e4c9ce210991c6edcb5370cb9b11e9e554a0f563e7754a4028a8fd0cb73",
      snapshot = "20171101T160520Z",
      url = "http://snapshot.debian.org/archive",
  )

  # download packages needed to create base docker image layers
  dpkg_list(
      name = "package_bundle",
      packages = [
          # iptables and dependencies (from https://packages.debian.org/stretch/iptables)
          "iptables",
          "libip4tc0",
          "libip6tc0",
          "libxtables12",

          # openssh-client and dependencies (from https://packages.debian.org/stretch/openssh-client)
          "openssh-client",
          "zlib1g",
          "libssl1.0.2",

          # jq
          "jq",
          "libjq1",
          "libonig4",
      ],
      sources = [
          "@debian_stretch//file:Packages.json",
      ],
  )

def _docker_dependencies_debian_slim_image():
  container_pull(
    name = "debian_slim_container_image",
    registry = "registry.hub.docker.com",
    repository = "library/debian",
    # stable-slim tag as of 9/23/2018
    digest = "sha256:76e4d780ebdd81315c1d67e0a044fabc06db5805352e3322594360d3990be1b6"
  )

def _docker_dependencies_helm():
  # download terraform binary to include in base docker image layer
  native.new_http_archive(
      name = "helm_bin",
      url = "https://storage.googleapis.com/kubernetes-helm/helm-v2.10.0-linux-amd64.tar.gz",
      sha256 = "0fa2ed4983b1e4a3f90f776d08b88b0c73fd83f305b5b634175cb15e61342ffe",
      build_file_content = """
filegroup(
    name = "bin",
    srcs = ["linux-amd64/helm"],
    visibility = ["//visibility:public"],
)
"""
  )
def _docker_dependencies_kubectl():
  # download terraform binary to include in base docker image layer
  native.new_http_archive(
      name = "kubectl_bin",
      url = "https://dl.k8s.io/v1.10.1/kubernetes-client-linux-amd64.tar.gz",
      sha256 = "f638cf6121e25762e2f6f36bca9818206778942465f0ea6e3ba59cfcc9c2738a",
      build_file_content = """
filegroup(
    name = "bin",
    srcs = ["kubernetes/client/bin/kubectl"],
    visibility = ["//visibility:public"],
)
"""
  )


def _docker_dependencies_terraform():
  # download terraform binary to include in base docker image layer
  native.new_http_archive(
      name = "terraform_bin",
      url = "https://releases.hashicorp.com/terraform/0.10.8/terraform_0.10.8_linux_amd64.zip",
      sha256 = "b786c0cf936e24145fad632efd0fe48c831558cc9e43c071fffd93f35e3150db",
      build_file_content = """
filegroup(
    name = "bin",
    srcs = ["terraform"],
    visibility = ["//visibility:public"],
)
"""
  )
