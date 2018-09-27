load(":image.bzl", "mock_go_container_image")

mock_go_container_image(
    name = "api-server",
    base_image = "//docker:go-openssh-client",
    path = "cmd/mock/api-server/rest",
)
