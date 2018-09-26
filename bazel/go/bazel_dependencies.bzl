def rules_go_dependencies():
  native.http_archive(
      name = "io_bazel_rules_go",
      url = "https://github.com/bazelbuild/rules_go/releases/download/0.15.3/rules_go-0.15.3.tar.gz",
      sha256 = "97cf62bdef33519412167fd1e4b0810a318a7c234f5f8dc4f53e2da86241c492",
  )

def bazel_gazelle_dependencies():
  native.http_archive(
      name = "bazel_gazelle",
      url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.14.0/bazel-gazelle-0.14.0.tar.gz",
      sha256 = "c0a5739d12c6d05b6c1ad56f2200cb0b57c5a70e03ebd2f7b87ce88cabf09c7b",
  )
