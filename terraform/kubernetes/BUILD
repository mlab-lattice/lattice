load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

# TODO: figure out why you need to do a bazel clean for this target to be updated
pkg_tar(
    name = "modules-tar",
    srcs = [":modules"],
    visibility = ["//visibility:public"],
)
