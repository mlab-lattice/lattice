load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar", "pkg_deb")
load("@io_bazel_rules_go//go:def.bzl", "go_binary")

_plugin_suffix = "-plugin"
_plugin_bin_suffix = "-plugin_bin"
_tar_suffix = "-tar"

def go_binary_docgen(
    name = "docs",
    output_file = "docs.md",
    embed = ":go_default_library",
    extra_markdown = None):
  plugin_name = name + _plugin_suffix
  plugin_bin_name = name + _plugin_bin_suffix

  go_binary(
      name = plugin_name,
      embed = [":go_default_library"],
      out = "plugin.so",
      linkmode = "plugin",
      visibility = ["//visibility:private"],
  )

  go_binary(
      name = plugin_bin_name,
      embed = ["//cmd/docgen:go_default_library"],
      data = [plugin_name],
      visibility = ["//visibility:private"],
  )

  cmd = "$(location {}) --plugin $(location {})".format(plugin_bin_name, plugin_name)
  srcs = [plugin_name, plugin_bin_name]
  outs = [output_file]
  if extra_markdown != None:
    cmd += " --extra-markdown={}".format(extra_markdown)
    srcs += ["//{}:extra-markdown".format(extra_markdown)]

  native.genrule(
      name = name,
      srcs = srcs,
      outs = outs,
      cmd = cmd + " > $@",
      visibility = ["//visibility:public"],
  )

  pkg_tar(
    name = name + _tar_suffix,
    srcs = outs,
    visibility = ["//visibility:public"],
)