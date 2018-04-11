load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

# Need to use a filegroup here so that bazel can keep track of the
# all of the files and which have changed.
# Was originally using
#  srcs = [":modules"]
# in pkg_tar, but bazel would not pick up that a file had changed
# so you would have to run bazel clean to rebuild targets that
# include the modules.
# more info: https://docs.bazel.build/versions/master/be/general.html#filegroup
filegroup(
    name = "modules-filegroup",
    srcs = glob(["modules/**/*.tf"]),
)

pkg_tar(
    name = "modules-tar",
    srcs = [":modules-filegroup"],

    # Without this, all of the .tf files are flattened into a single directory
    # more info: https://github.com/bazelbuild/bazel/issues/2176
    strip_prefix = ".",
    visibility = ["//visibility:public"],
)
