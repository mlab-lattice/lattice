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

# base container images
base_images = [
    "iptables",
    "terraform",
]
lattice_base_container_images(base_images)

# lattice container images
# this will yield the following rules:
# (container-image) foo
# (container-push) push-foo
# (container-image) debug-foo
# (container-push) push-debug-foo
go_targets = [
    ("envoy-prepare", "iptables", "cmd/envoy/prepare"),
    ("kubernetes-bootstrap-lattice", None, "cmd/kubernetes/bootstrap-lattice"),
    ("kubernetes-component-builder", None, "cmd/kubernetes/componentbuild"),
    ("kubernetes-envoy-xds-api-rest-per-node", None, "cmd/kubernetes/envoy/xds-api/rest/per-node"),
    ("kubernetes-lattice-controller-manager", "terraform", "cmd/kubernetes/lattice-controller-manager"),
    ("kubernetes-manager-api-rest", None, "cmd/kubernetes/manager/api-rest"),
]
lattice_container_images(go_targets)





#
## prod images
#container_image(
#    name = "iptables",
#    base = "@go_image_base//image",
#    tars = [":base-iptables"],
#)
#
#container_image(
#    name = "terraform",
#    base = "@go_image_base//image",
#    tars = [":base-terraform"],
#)
#
## debug images
#container_image(
#    name = "debug-terraform",
#    base = "@go_debug_image_base//image",
#    tars = [":base-terraform"],
#)
#
#container_image(
#    name = "debug-iptables",
#    base = "@go_debug_image_base//image",
#    tars = [":base-iptables"],
#)
## prod build/push
#go_image(
#    name = "envoy-prepare",
#    base = ":iptables",
#    importpath = "github.com/mlab-lattice/system/cmd/envoy/prepare",
#    library = "//cmd/envoy/prepare:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "envoy-prepare-push",
#    format = "Docker",
#    image = ":envoy-prepare",
#    registry = "gcr.io/lattice-dev",
#    repository = "envoy-prepare",
#)
#
#go_image(
#    name = "kubernetes-bootstrap-lattice",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/bootstrap-lattice",
#    library = "//cmd/kubernetes/bootstrap-lattice:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "kubernetes-bootstrap-lattice-push",
#    format = "Docker",
#    image = ":kubernetes-bootstrap-lattice",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-bootstrap-lattice",
#)
#
#go_image(
#    name = "kubernetes-component-builder",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/componentbuild",
#    library = "//cmd/kubernetes/componentbuild:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "kubernetes-component-builder-push",
#    format = "Docker",
#    image = ":kubernetes-component-builder",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-component-builder",
#)
#
#go_image(
#    name = "kubernetes-envoy-xds-api-rest-per-node",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/envoy/xds-api/rest/per-node",
#    library = "//cmd/kubernetes/envoy/xds-api/rest/per-node:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "kubernetes-envoy-xds-api-rest-per-node-push",
#    format = "Docker",
#    image = ":kubernetes-envoy-xds-api-rest-per-node",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-envoy-xds-api-rest-per-node",
#)
#
#go_image(
#    name = "kubernetes-lattice-controller-manager",
#    base = ":terraform",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager",
#    library = "//cmd/kubernetes/lattice-controller-manager:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "kubernetes-lattice-controller-manager-push",
#    format = "Docker",
#    image = ":kubernetes-lattice-controller-manager",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-lattice-controller-manager",
#)
#
#go_image(
#    name = "kubernetes-manager-api-rest",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/manager/api-rest",
#    library = "//cmd/kubernetes/manager/api-rest:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "kubernetes-manager-api-rest-push",
#    format = "Docker",
#    image = ":kubernetes-manager-api-rest",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-manager-api-rest",
#)
#
## debug build/push
#go_image(
#    name = "debug-envoy-prepare",
#    base = "debug-iptables",
#    importpath = "github.com/mlab-lattice/system/cmd/envoy/prepare",
#    library = "//cmd/envoy/prepare:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-envoy-prepare-push",
#    format = "Docker",
#    image = ":debug-envoy-prepare",
#    registry = "gcr.io/lattice-dev",
#    repository = "debug-envoy-prepare",
#)
#
#go_image(
#    name = "debug-kubernetes-bootstrap-lattice",
#    base = "@go_debug_image_base//image",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/bootstrap-lattice",
#    library = "//cmd/kubernetes/bootstrap-lattice:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-kubernetes-bootstrap-lattice-push",
#    format = "Docker",
#    image = ":debug-kubernetes-bootstrap-lattice",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-bootstrap-lattice",
#)
#
#go_image(
#    name = "debug-kubernetes-component-builder",
#    base = "@go_debug_image_base//image",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/componentbuild",
#    library = "//cmd/kubernetes/componentbuild:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-kubernetes-component-builder-push",
#    format = "Docker",
#    image = ":debug-kubernetes-component-builder",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-component-builder",
#)
#
#go_image(
#    name = "debug-kubernetes-envoy-xds-api-rest-per-node",
#    base = "@go_debug_image_base//image",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/envoy/xds-api/rest/per-node",
#    library = "//cmd/kubernetes/envoy/xds-api/rest/per-node:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-kubernetes-envoy-xds-api-rest-per-node-push",
#    format = "Docker",
#    image = ":debug-kubernetes-envoy-xds-api-rest-per-node",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-envoy-xds-api-rest-per-node",
#)
#
#go_image(
#    name = "debug-kubernetes-lattice-controller-manager",
#    base = ":debug-terraform",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager",
#    library = "//cmd/kubernetes/lattice-controller-manager:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-kubernetes-lattice-controller-manager-push",
#    format = "Docker",
#    image = ":debug-kubernetes-lattice-controller-manager",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-lattice-controller-manager",
#)
#
#go_image(
#    name = "debug-kubernetes-manager-api-rest",
#    base = "@go_debug_image_base//image",
#    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/manager/api-rest",
#    library = "//cmd/kubernetes/manager/api-rest:go_default_library",
#    visibility = ["//visibility:public"],
#)
#
#container_push(
#    name = "debug-kubernetes-manager-api-rest-push",
#    format = "Docker",
#    image = ":debug-kubernetes-manager-api-rest",
#    registry = "gcr.io/lattice-dev",
#    repository = "kubernetes-manager-api-rest",
#)
