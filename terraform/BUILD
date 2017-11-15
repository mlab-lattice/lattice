load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

pkg_tar(
    name = "modules-tar",
    srcs = [":modules"],
    package_dir = "/foo",
    visibility = ["//visibility:public"],
)

