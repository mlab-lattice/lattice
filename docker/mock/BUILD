load(":image.bzl", "mock_go_container_image")

mock_go_container_image(
    name = "api-server-rest",
    base_image = "//docker:go-openssh-client",
    path = "cmd/backend/kubernetes/api-server/rest",
)
