load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("//bazel/docker:image.bzl", "lattice_base_go_container_image", "lattice_go_container_image")
load("@package_bundle//file:packages.bzl", "packages")

# base container image layers
container_image(
    name = "base-iptables",
    debs = [
        # See bazel/docker/dependencies.bzl for more info.
        packages["iptables"],
        packages["libip4tc0"],
        packages["libip6tc0"],
        packages["libxtables12"],
    ],
    visibility = ["//visibility:public"],
)

container_image(
    name = "base-openssh-client",
    debs = [
        # See bazel/docker/dependencies.bzl for more info.
        packages["openssh-client"],
        packages["zlib1g"],
    ],
    visibility = ["//visibility:public"],
)

container_image(
    name = "base-terraform-bin",
    directory = "/usr/local/bin",
    files = ["@terraform_bin//:bin"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "base-terraform-modules-kubernetes",
    directory = "/etc/terraform",
    tars = ["//terraform/kubernetes:modules-tar"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "base-terraform",
    tars = [
        ":base-terraform-bin",
        ":base-terraform-modules-kubernetes",
    ],
    visibility = ["//visibility:public"],
)

# base container go images
lattice_base_go_container_image(
    name = "iptables",
    base_layer_target = ":base-iptables",
)

lattice_base_go_container_image(
    name = "openssh-client",
    base_layer_target = ":base-openssh-client",
)

lattice_base_go_container_image(
    name = "terraform",
    base_layer_target = ":base-terraform",
)

# lattice go container images
lattice_go_container_image(
    name = "kubernetes-api-server-rest",
    base_image = "openssh-client",
    path = "cmd/kubernetes/api-server/rest",
)

lattice_go_container_image(
    name = "kubernetes-container-builder",
    base_image = "openssh-client",
    path = "cmd/kubernetes/container-builder",
)

lattice_go_container_image(
    name = "kubernetes-envoy-prepare",
    base_image = "iptables",
    path = "cmd/kubernetes/envoy/prepare",
)

lattice_go_container_image(
    name = "kubernetes-envoy-xds-api-grpc-per-node",
    base_image = None,
    path = "cmd/kubernetes/envoy/xdsapi/v2/grpc/per-node",
)

lattice_go_container_image(
    name = "kubernetes-lattice-controller-manager",
    base_image = "terraform",
    path = "cmd/kubernetes/lattice-controller-manager",
)

lattice_go_container_image(
    name = "kubernetes-local-dns-controller",
    base_image = None,
    path = "cmd/kubernetes/cloud-provider/local/dns-controller",
)

lattice_go_container_image(
    name = "latticectl",
    base_image = None,
    path = "cmd/latticectl",
)

lattice_go_container_image(
    name = "mock-api-server",
    base_image = None,
    path = "cmd/mock/api-server",
)
