load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_base_images = {
    False: "@go_image_base//image",
    True: "@go_debug_image_base//image",
}

debug_prefix = "debug-"
push_prefix = "push-"

# creates both production and debug versions of the base go image
# the base_image argument should be the name of a container_image target
def lattice_base_go_container_image(name, base_layer_target):
  _lattice_base_go_container_image(
      name=name,
      base_layer_target=base_layer_target,
      debug=False,
  )

  _lattice_base_go_container_image(
      name=name,
      base_layer_target=base_layer_target,
      debug=True,
  )

def _lattice_base_go_container_image(name, base_layer_target, debug=False):
  name = name if not debug else debug_prefix + name
  container_image(
      name = name,
      base = go_base_images[debug],
      tars = [base_layer_target],
  )


# creates both go_image and container_push targets for both stripped and debug versions
# of the target
def lattice_go_container_image(name, base_image, path):
  debug_name = debug_prefix + name

  prod_base_image = base_image if base_image else go_base_images[False]
  debug_base_image = debug_prefix + base_image if base_image else go_base_images[True]

  go_image(
      name = name,
      base = prod_base_image,
      embed = ["//" + path + ":go_default_library"],
      visibility = ["//visibility:public"],
  )

  container_push(
      name = push_prefix + name,
      format = "Docker",
      image = ":" + name,
      registry = "{REGISTRY}",
      repository = "{CHANNEL}-" + name,
      stamp = True,
  )

  go_image(
      name = debug_name,
      base = debug_base_image,
      embed = ["//" + path + ":go_default_library"],
      visibility = ["//visibility:public"],
  )

  container_push(
      name = push_prefix + debug_name,
      format = "Docker",
      image = ":" + debug_name,
      registry = "{REGISTRY}",
      repository = "{CHANNEL}-" + debug_name,
      stamp = True,
  )