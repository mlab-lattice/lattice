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
def lattice_base_go_container_image(name, base):
  _lattice_base_go_container_image(
      name=name,
      base=base,
      stripped=False,
  )

  _lattice_base_go_container_image(
      name=name,
      base=base,
      stripped=True,
  )

def _lattice_base_go_container_image(name, base, stripped=False):
  name = name if not stripped else name + stripped_target_suffix
  container_image(
      name = name,
      base = go_base_images[stripped],
      tars = [base],
      visibility = ["//visibility:public"],
  )


# creates a container_push target for the image
def lattice_container_push(image, image_name):
  container_push(
      name = push_target_prefix + image,
      format = "Docker",
      image = image,
      registry = "{REGISTRY}",
      repository = "{REPOSITORY_PREFIX}/{CHANNEL}/" + image_name,
      stamp = True,
  )


# creates a container_image and lattice_container_push for the given name, image_name,
# and container_image arguments
def lattice_container_image(name, image_name, **kwargs):
  container_image(
      name = name,
      **kwargs
  )

  lattice_container_push(
      image = name,
      image_name = image_name,
  )

# creates both go_image and lattice_container_push targets for both normal and stripped versions
# of the target
def lattice_go_container_image(name, image_name, base_image, path):
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

  lattice_container_push(
      image = name,
      image_name = image_name,
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

  lattice_container_push(
      image = stripped_name,
      image_name = "stripped/" + image_name,
  )
