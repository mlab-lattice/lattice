load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_base_images = {
    False: "@go_debug_image_base//image",
    True: "@go_image_base//image",
}

push_target_prefix = "push-"
stripped_target_suffix = "-stripped"

# creates both normal and stripped versions of the base go image
# the base_image argument should be the name of a container_image target
def lattice_base_go_container_image(name, base_layer_target):
  _lattice_base_go_container_image(
      name=name,
      base_layer_target=base_layer_target,
      stripped=False,
  )

  _lattice_base_go_container_image(
      name=name,
      base_layer_target=base_layer_target,
      stripped=True,
  )

def _lattice_base_go_container_image(name, base_layer_target, stripped=False):
  name = name if not stripped else name + stripped_target_suffix
  container_image(
      name = name,
      base = go_base_images[stripped],
      tars = [base_layer_target],
  )


# creates both go_image and container_push targets for both normal and stripped versions
# of the target
def lattice_go_container_image(name, base_image, path):
  stripped_name = name + stripped_target_suffix

  stripped_base_image = base_image + stripped_target_suffix if base_image else go_base_images[True]
  base_image = base_image if base_image else go_base_images[False]

  go_image(
      name = name,
      base = base_image,
      embed = ["//" + path + ":go_default_library"],
      goos = "linux",
      goarch = "amd64",
      pure = "on",
      visibility = ["//visibility:public"],
  )

  container_push(
      name = push_target_prefix + name,
      format = "Docker",
      image = ":" + name,
      registry = "{REGISTRY}",
      repository = "{REPOSITORY_PREFIX}/{CHANNEL}/" + name,
      stamp = True,
  )

  go_image(
      name = stripped_name,
      base = stripped_base_image,
      embed = ["//" + path + ":go_default_library"],
      goos = "linux",
      goarch = "amd64",
      pure = "on",
      visibility = ["//visibility:public"],
  )

  container_push(
      name = push_target_prefix + stripped_name,
      format = "Docker",
      image = ":" + stripped_name,
      registry = "{REGISTRY}",
      repository = "{REPOSITORY_PREFIX}/{CHANNEL}/stripped/" + name,
      stamp = True,
  )