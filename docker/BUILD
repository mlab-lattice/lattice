load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load(":image.bzl", "lattice_base_go_container_image")
load("@package_bundle//file:packages.bzl", "packages")

container_image(
    name = "base-openssh-client",
    debs = [
        # See bazel/docker/dependencies.bzl for more info.
        packages["openssh-client"],
        packages["zlib1g"],
    ],
)

lattice_base_go_container_image(
    name = "go-openssh-client",
    base_layer_target = ":base-openssh-client",
)