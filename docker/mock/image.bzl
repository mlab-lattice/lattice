load("//docker:image.bzl", "lattice_container_image", "lattice_go_container_image")

_mock_image_name_prefix = "mock/"

def mock_go_container_image(name, base_image, path):
  lattice_go_container_image(
      name = name,
      image_name = _mock_image_name_prefix + name,
      base_image = base_image,
      path = path,
  )
