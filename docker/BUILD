load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@package_bundle//file:packages.bzl", "packages")

# base images
container_image(
    name = "iptables",
    base = "@distroless_base//image",
    debs = [
        # See WORKSPACE for more info.
        packages["iptables"],
        packages["libip4tc0"],
        packages["libip6tc0"],
        packages["libxtables12"],
    ],
)

container_image(
    name = "terraform-bin",
    directory = "/usr/local/bin",
    files = ["@terraform_bin//:bin"],
)

container_image(
    name = "terraform-modules-kubernetes",
    directory = "/etc/terraform",
    tars = ["//terraform/kubernetes:modules-tar"],
)

container_image(
    name = "terraform",
    base = "@go_image_base//image",
    tars = [
        ":terraform-bin",
        ":terraform-modules-kubernetes",
    ]
)

# actual image/push
container_image(
    name = "envoy-prepare-envoy",
    base = ":iptables",
    directory = "/usr/local/bin",
    files = ["//bin/envoy:prepare-envoy"],
    entrypoint = "/usr/local/bin/prepare-envoy",
)

container_push(
    name = "envoy-prepare-envoy-push",
    format = "Docker",
    image = ":envoy-prepare-envoy",
    registry = "gcr.io/lattice-dev",
    repository = "envoy-prepare-envoy",
)

go_image(
    name = "kubernetes-bootstrap-lattice",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/bootstrap-lattice",
    library = "//cmd/kubernetes/bootstrap-lattice:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "kubernetes-bootstrap-lattice-push",
    format = "Docker",
    image = ":kubernetes-bootstrap-lattice",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-bootstrap-lattice",
)

go_image(
    name = "kubernetes-component-builder",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/componentbuild",
    library = "//cmd/kubernetes/componentbuild:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "kubernetes-component-builder-push",
    format = "Docker",
    image = ":kubernetes-component-builder",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-component-builder",
)

go_image(
    name = "kubernetes-envoy-xds-api-rest-per-node",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/envoy/xds-api/rest/per-node",
    library = "//cmd/kubernetes/envoy/xds-api/rest/per-node:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "kubernetes-envoy-xds-api-rest-per-node-push",
    format = "Docker",
    image = ":kubernetes-envoy-xds-api-rest-per-node",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-envoy-xds-api-rest-per-node",
)

go_image(
    name = "kubernetes-lattice-controller-manager",
    base = ":terraform",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager",
    library = "//cmd/kubernetes/lattice-controller-manager:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "kubernetes-lattice-controller-manager-push",
    format = "Docker",
    image = ":kubernetes-lattice-controller-manager",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-lattice-controller-manager",
)

go_image(
    name = "kubernetes-manager-api-rest",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/manager/api-rest",
    library = "//cmd/kubernetes/manager/api-rest:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "kubernetes-manager-api-rest-push",
    format = "Docker",
    image = ":kubernetes-manager-api-rest",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-manager-api-rest",
)