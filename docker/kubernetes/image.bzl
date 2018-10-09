load("//docker:image.bzl", "lattice_container_image", "lattice_go_container_image")

_kubernetes_image_name_prefix = "kubernetes/"

def kubernetes_container_image(name, image_name, **kwargs):
  lattice_container_image(
      name = name,
      image_name = _kubernetes_image_name_prefix + image_name,
      **kwargs
  )

def kubernetes_go_container_image(name, image_name, base_image, path):
  lattice_go_container_image(
      name = name,
      image_name = _kubernetes_image_name_prefix + image_name,
      base_image = base_image,
      path = path,
  )
