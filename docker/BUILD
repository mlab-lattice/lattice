load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load(":image.bzl", "lattice_base_go_container_image")
load("@package_bundle//file:packages.bzl", "packages")

container_image(
    name = "base-openssh-client",
    debs = [
        # See bazel/docker/dependencies.bzl for more info.
        packages["openssh-client"],
        packages["zlib1g"],
        packages["libssl1.0.2"],
    ],
    visibility = ["//visibility:public"],
)

lattice_base_go_container_image(
    name = "go-openssh-client",
    base = ":base-openssh-client",
)
