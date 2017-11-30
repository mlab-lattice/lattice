load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")

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
    name = "terraform-modules-kubernetes",
    base = ":terraform",
    directory = "/etc/terraform",
    tars = ["//terraform/kubernetes:modules-tar"],
)

container_image(
    name = "wrap-with-ssh-agent",
    base = "@debian_with_ssh//image",
    directory = "/usr/local/bin",
    files = ["//bin/ssh:wrap-with-ssh-agent"],
    entrypoint = "/usr/local/bin/wrap-with-ssh-agent",
)

container_image(
    name = "wrap-with-ssh-agent-and-known-hosts",
    base = ":wrap-with-ssh-agent",
    directory = "/etc/ssh",
    files = ["//docker/kubernetes/manager/api/ssh:ssh_known_hosts"],
)

load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "cli-user",
    base = ":terraform-modules-kubernetes",
    importpath = "github.com/mlab-lattice/system/cmd/cli/user",
    library = "//cmd/cli/user:go_default_library",
    visibility = ["//visibility:public"],
)

container_push(
    name = "cli-user-push",
    format = "Docker",
    image = ":cli-user",
    registry = "gcr.io/lattice-dev",
    repository = "lattice-cli-user",
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
    base = ":terraform-modules-kubernetes",
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

container_image(
    name = "kubernetes-manager-api-rest",
    base = ":wrap-with-ssh-agent-and-known-hosts",
    directory = "/app/docker",
    files = ["//cmd/kubernetes/manager/api-rest"],
    entrypoint = [
        "/usr/local/bin/wrap-with-ssh-agent",
        "/app/docker/api-rest"
    ],
)

container_push(
    name = "kubernetes-manager-api-rest-push",
    format = "Docker",
    image = ":kubernetes-manager-api-rest",
    registry = "gcr.io/lattice-dev",
    repository = "kubernetes-manager-api-rest",
)