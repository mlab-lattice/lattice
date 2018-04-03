load("@distroless//package_manager:package_manager.bzl", "dpkg_src", "dpkg_list")

def docker_dependencies():
  _docker_dependencies_debian_pkg()
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
          # libstdc++6 and dependencies (from https://packages.debian.org/stretch/libstdc%2B%2B6)
          # needed for admin CLI now for some reason
          "libstdc++6",
          "libgcc1",

          # iptables and dependencies (from https://packages.debian.org/stretch/iptables)
          "iptables",
          "libip4tc0",
          "libip6tc0",
          "libxtables12",

          # openssh-client and dependencies (from https://packages.debian.org/stretch/openssh-client)
          "openssh-client",
          "zlib1g",
      ],
      sources = [
          "@debian_stretch//file:Packages.json",
      ],
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
