load("@io_bazel_rules_docker//container:container.bzl", "container_image")

container_image(
    name = "component-build-build-docker-image",
    base = "@docker_git//image",
    directory = "/usr/local/bin",
    files = ["//bin/component-build:build-docker-image"],
    entrypoint = "/usr/local/bin/build-docker-image",
)

container_image(
    name = "component-build-get-ecr-creds",
    base = "@ubuntu_with_aws//image",
    directory = "/usr/local/bin",
    files = ["//bin/component-build:get-ecr-creds"],
    entrypoint = "/usr/local/bin/get-ecr-creds",
)

container_image(
    name = "component-build-pull-git-repo",
    base = "@docker_git//image",
    directory = "/usr/local/bin",
    files = ["//bin/component-build:pull-git-repo"],
    entrypoint = "/usr/local/bin/pull-git-repo",
)

container_image(
    name = "envoy-prepare-envoy",
    base = "@ubuntu_with_iptables//image",
    directory = "/usr/local/bin",
    files = ["//bin/envoy:prepare-envoy"],
    entrypoint = "/usr/local/bin/prepare-envoy",
)

container_image(
    name = "terraform",
    base = "@go_image_base//image",
    directory = "/usr/local/bin",
    files = ["@terraform_bin//:bin"],
)

container_image(
    name = "terraform-modules",
    tars = ["//terraform:modules-tar"],
    base = ":terraform",
)

load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "lattice-system-cli",
    importpath = "github.com/mlab-lattice/system/cmd/cli",
    library = "//cmd/cli:go_default_library",
    visibility = ["//visibility:public"],
)

go_image(
    name = "kubernetes-bootstrap-lattice",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/bootstrap-lattice",
    library = "//cmd/kubernetes/bootstrap-lattice:go_default_library",
    visibility = ["//visibility:public"],
)

go_image(
    name = "kubernetes-envoy-xds-api-rest-per-node",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/envoy/xds-api/rest/per-node",
    library = "//cmd/kubernetes/envoy/xds-api/rest/per-node:go_default_library",
    visibility = ["//visibility:public"],
)

go_image(
    name = "kubernetes-lattice-controller-manager",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager",
    library = "//cmd/kubernetes/lattice-controller-manager:go_default_library",
    visibility = ["//visibility:public"],
    base = ":terraform-modules",
)

go_image(
    name = "kubernetes-manager-api-rest",
    importpath = "github.com/mlab-lattice/system/cmd/kubernetes/manager/api-rest",
    library = "//cmd/kubernetes/manager/api-rest:go_default_library",
    visibility = ["//visibility:public"],
)
