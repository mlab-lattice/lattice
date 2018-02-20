load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_base_images = {
    False: "@go_image_base//image",
    True: "@go_debug_image_base//image",
}

build_user_stamp_prefix = "{BUILD_USER}-"
debug_prefix = "debug-"
push_prefix = "push-"
registry = "gcr.io/lattice-dev"
stable_prefix = "stable-"
user_prefix = "user-"

def lattice_base_container_images(base_images):
  for image in base_images:
    lattice_base_container_image(image, False)
    lattice_base_container_image(image, True)


def lattice_base_container_image(base_image, debug=False):
  name = base_image if not debug else debug_prefix + base_image
  base = "@go_image_base//image"
  container_image(
      name = name,
      base = go_base_images[debug],
      tars = [":base-" + base_image],
  )


def lattice_container_images(go_targets):
  for target in go_targets:
    lattice_go_container_image(target)


def lattice_go_container_image(target, debug=False):
  (name, base_image, path) = target

  debug_name = debug_prefix + name

  prod_base_image = base_image if base_image else go_base_images[False]
  debug_base_image = debug_prefix + base_image if base_image else go_base_images[True]

  go_image(
      name = name,
      base = prod_base_image,
      embed = ["//" + path + ":go_default_library"],
      visibility = ["//visibility:public"],
      pure = "on",
#      goos = "linux",
#      goarch = "amd64",
#      pure = "on",
  )

  container_push(
      name = push_prefix + stable_prefix + name,
      format = "Docker",
      image = ":" + name,
      registry = registry,
      repository = stable_prefix + name,
  )

  container_push(
      name = push_prefix + user_prefix + name,
      format = "Docker",
      image = ":" + name,
      registry = registry,
      repository = build_user_stamp_prefix + name,
      stamp = True,
  )

  go_image(
      name = debug_name,
      base = debug_base_image,
      embed = ["//" + path + ":go_default_library"],
      visibility = ["//visibility:public"],
      pure = "on",
  )


  container_push(
      name = push_prefix + stable_prefix + debug_name,
      format = "Docker",
      image = ":" + debug_name,
      registry = registry,
      repository = stable_prefix + debug_name,
  )


  container_push(
      name = push_prefix + user_prefix + debug_name,
      format = "Docker",
      image = ":" + debug_name,
      registry = registry,
      repository = build_user_stamp_prefix + debug_name,
      stamp = True,
  )