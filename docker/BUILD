load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load(":lattice_targets.bzl", "lattice_base_container_images", "lattice_container_images")
load("@package_bundle//file:packages.bzl", "packages")

# base container image layers
container_image(
    name = "base-iptables",
    debs = [
        # See WORKSPACE for more info.
        packages["iptables"],
        packages["libip4tc0"],
        packages["libip6tc0"],
        packages["libxtables12"],
    ],
)

container_image(
    name = "base-libstdc++6",
    debs = [
        # See bazel/dependencies.bzl for more info.
        packages["libstdc++6"],
        packages["libgcc1"],
    ],
)

container_image(
    name = "base-terraform-bin",
    directory = "/usr/local/bin",
    files = ["@terraform_bin//:bin"],
)

container_image(
    name = "base-terraform-modules-kubernetes",
    directory = "/etc/terraform",
    tars = ["//terraform/kubernetes:modules-tar"],
)

container_image(
    name = "base-terraform",
    tars = [
        ":base-terraform-bin",
        ":base-terraform-modules-kubernetes",
    ]
)

container_image(
    name = "base-libstdc++6-terraform",
    tars = [
        ":base-libstdc++6",
        ":base-terraform",
    ]
)

# base container images
base_images = [
    "iptables",
    "libstdc++6",
    "libstdc++6-terraform",
]
lattice_base_container_images(base_images)

# lattice container images
# this will yield the following rules:
# (container-image) foo
# (container-image) debug-foo
# (container-push) push-foo
# (container-push) push-debug-foo
go_targets = [
    ("envoy-prepare", "iptables", "cmd/envoy/prepare"),
    ("kubernetes-component-builder", "libstdc++6", "cmd/kubernetes/componentbuilder"),
    ("kubernetes-envoy-xds-api-rest-per-node", "libstdc++6", "cmd/kubernetes/envoy/xdsapi/v1/rest/per-node"),
    ("kubernetes-lattice-controller-manager", "libstdc++6-terraform", "cmd/kubernetes/lattice-controller-manager"),
    ("kubernetes-manager-api-rest", "libstdc++6", "cmd/kubernetes/manager-api/rest"),
    ("kubernetes-local-dns-controller", "libstdc++6", "cmd/kubernetes/local-dns"),
    ("latticectl", "libstdc++6", "cmd/latticectl"),
]
lattice_container_images(go_targets)
