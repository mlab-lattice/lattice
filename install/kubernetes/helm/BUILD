load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

filegroup(
    name = "filegroup",
    srcs = glob(["**/*"]),
)

pkg_tar(
    name = "tarball",
    srcs = [":filegroup"],

    # Without this, all of the files are flattened into a single directory
    # more info: https://github.com/bazelbuild/bazel/issues/2176
    strip_prefix = ".",
    visibility = ["//visibility:public"],
)
