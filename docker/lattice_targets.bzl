load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_base_images = {
    False: "@go_image_base//image",
    True: "@go_debug_image_base//image",
}


def lattice_base_container_images(base_images):
  for image in base_images:
    lattice_base_container_image(image, False)
    lattice_base_container_image(image, True)


def lattice_base_container_image(base_image, debug=False):
  name = base_image if not debug else "debug-" + base_image
  base = "@go_image_base//image"
  container_image(
      name = name,
      base = go_base_images[debug],
      tars = [":base-" + base_image],
  )


def lattice_container_images(go_targets):
  for target in go_targets:
    lattice_go_container_image(target, False)
    lattice_go_container_image(target, True)


def lattice_go_container_image(target, debug=False):
  (name, base_image, path) = target
  name = name if not debug else "debug-" + name

  if base_image:
    base_image = base_image if not debug else "debug-" + base_image
  else:
    base_image = go_base_images[debug]

  go_image(
      name = name,
      base = base_image,
      importpath = "github.com/mlab-lattice/system/" + path,
      library = "//" + path + ":go_default_library",
      visibility = ["//visibility:public"],
  )

  container_push(
      name = "push-" + name,
      format = "Docker",
      image = ":" + name,
      registry = "gcr.io/lattice-dev",
      repository = name,
  )