load("@io_bazel_rules_go//go:def.bzl", "go_binary")

_plugin_suffix = "_plugin"
_plugin_bin_suffix = "_plugin_bin"

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
  if extra_markdown != None:
    cmd += " --extra-markdown={}".format(extra_markdown)
    srcs += ["//{}:extra-markdown".format(extra_markdown)]

  native.genrule(
      name = name,
      srcs = srcs,
      outs = [output_file],
      cmd = cmd + " > $@",
      visibility = ["//visibility:public"],
  )
