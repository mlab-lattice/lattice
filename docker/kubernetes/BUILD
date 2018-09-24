load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("//docker:image.bzl", "lattice_base_go_container_image")
load(":image.bzl", "kubernetes_container_image", "kubernetes_go_container_image")
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
)

container_image(
    name = "base-helm-bin",
    directory = "/usr/local/bin",
    files = ["@helm_bin//:bin"],
)

container_image(
    name = "base-kubectl-bin",
    directory = "/usr/local/bin",
    files = ["@kubectl_bin//:bin"],
)

container_image(
    name = "base-kubernetes-installer-helm",
    base = "@debian_slim_container_image//image",
    tars = [
        ":base-helm-bin",
        ":base-kubectl-bin",
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
    ],
)

# base container go images
lattice_base_go_container_image(
    name = "go-iptables",
    base = ":base-iptables",
)

lattice_base_go_container_image(
    name = "go-terraform",
    base = ":base-terraform",
)

# go images
kubernetes_go_container_image(
    name = "api-server",
    image_name = "api-server",
    base_image = "//docker:go-openssh-client",
    path = "cmd/kubernetes/api-server/rest",
)

kubernetes_go_container_image(
    name = "container-builder",
    image_name = "container-builder",
    base_image = "//docker:go-openssh-client",
    path = "cmd/kubernetes/container-builder",
)

kubernetes_go_container_image(
    name = "envoy-prepare",
    image_name = "envoy/prepare",
    base_image = ":go-iptables",
    path = "cmd/kubernetes/envoy/prepare",
)

kubernetes_go_container_image(
    name = "envoy-xds-api",
    image_name = "envoy/xds-api",
    base_image = None,
    path = "cmd/kubernetes/envoy/xdsapi/v2/grpc/per-node",
)

kubernetes_go_container_image(
    name = "controller-manager",
    image_name = "controller-manager",
    base_image = ":go-terraform",
    path = "cmd/kubernetes/controller-manager",
)

kubernetes_go_container_image(
    name = "local-dns-controller",
    image_name = "local/dns-controller",
    base_image = None,
    path = "cmd/kubernetes/local/dns-controller",
)

# installer image
kubernetes_container_image(
    name = "installer-helm",
    image_name = "installer/helm",
    base = ":base-kubernetes-installer-helm",
    directory = "/etc/lattice/install/kubernetes/helm",
    tars = ["//install/kubernetes/helm:tarball"],
    entrypoint = "/etc/lattice/install/kubernetes/helm/install.sh",
)
