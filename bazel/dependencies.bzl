load("@io_bazel_rules_go//go:def.bzl", "go_repository")
load(":bazel/go_repositories.bzl", "GO_REPOSITORIES")

def go_dependencies():
    dependencies = [
        "github.com/aws/aws-sdk-go",
        "github.com/coreos/go-iptables",
        "github.com/docker/docker",
        "github.com/fatih/color",
        "github.com/gin-gonic/gin",
        "github.com/satori/go.uuid",
        "github.com/sergi/go-diff",
        "github.com/spf13/cobra",
        "github.com/olekukonko/tablewriter",
        "golang.org/x/crypto",
        "gopkg.in/src-d/go-git.v4",
        "k8s.io/api",
        "k8s.io/apiextensions-apiserver",
        "k8s.io/apimachinery",
        "k8s.io/client-go",
        "k8s.io/kubernetes",
    ]

    for dep in dependencies:
      go_repository(**GO_REPOSITORIES[dep])

    _go_dependencies_com_github_aws_aws_sdk_go()
    _go_dependencies_com_github_docker_docker()
    _go_dependencies_com_github_fatih_color()
    _go_dependencies_com_github_gin_gonic_gin()
    _go_dependencies_com_github_olekukonko_tablewriter()
    _go_dependencies_com_github_spf13_cobra()
    _go_dependencies_in_gopkg_src_d_go_git_v4()
    _go_dependencies_io_k8s()

def _go_dependencies_com_github_aws_aws_sdk_go():
  dependencies = [
      "github.com/go-ini/ini",
      "github.com/jmespath/go-jmespath",
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_com_github_docker_docker():
  dependencies = [
      "github.com/docker/distribution",
      "github.com/docker/go-connections",
      "github.com/docker/go-units",
      "github.com/opencontainers/runc",
      "github.com/pkg/errors",
      "github.com/Sirupsen/logrus",
      "github.com/opencontainers/go-digest",
      "github.com/Nvveen/Gotty",
      "github.com/docker/libtrust",
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_com_github_fatih_color():
  dependencies = [
      "github.com/mattn/go-colorable"
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_com_github_gin_gonic_gin():
  dependencies = [
      "github.com/gin-contrib/sse",
      "github.com/mattn/go-isatty",
      "gopkg.in/go-playground/validator.v8",
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_com_github_olekukonko_tablewriter():
  dependencies = [
      "github.com/mattn/go-runewidth",
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_com_github_spf13_cobra():
  dependencies = [
      "github.com/spf13/pflag",
  ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_in_gopkg_src_d_go_git_v4():
    dependencies = [
      "github.com/jbenet/go-context",
      "github.com/mitchellh/go-homedir",
      "github.com/src-d/gcfg",
      "github.com/xanzy/ssh-agent",
      "gopkg.in/src-d/go-billy.v3",
      "gopkg.in/warnings.v0",
    ]

    for dep in dependencies:
      go_repository(**GO_REPOSITORIES[dep])

def _go_dependencies_io_k8s():
  dependencies = [
       "github.com/PuerkitoBio/purell",
       "github.com/PuerkitoBio/urlesc",
       "github.com/davecgh/go-spew",
       "github.com/emicklei/go-restful",
       "github.com/emicklei/go-restful-swagger12",
       "github.com/ghodss/yaml",
       "github.com/go-openapi/jsonpointer",
       "github.com/go-openapi/jsonreference",
       "github.com/go-openapi/spec",
       "github.com/go-openapi/swag",
       "github.com/gogo/protobuf",
       "github.com/google/btree",
       "github.com/google/gofuzz",
       "github.com/googleapis/gnostic",
       "github.com/gregjones/httpcache",
       "github.com/hashicorp/golang-lru",
       "github.com/howeyc/gopass",
       "github.com/imdario/mergo",
       "github.com/json-iterator/go",
       "github.com/juju/ratelimit",
       "github.com/mailru/easyjson",
       "github.com/pborman/uuid",
       "github.com/peterbourgon/diskv",
       "github.com/ugorji/go",
       "golang.org/x/sys",
       "gopkg.in/inf.v0",
       "gopkg.in/yaml.v2",
       "k8s.io/apiserver",
       "k8s.io/kube-openapi",
   ]

  for dep in dependencies:
    go_repository(**GO_REPOSITORIES[dep])

load("@distroless//package_manager:package_manager.bzl", "dpkg_src", "dpkg_list")

def docker_dependencies():
  _docker_dependencies_debian_pkg()
  _docker_dependencies_terraform()

def _docker_dependencies_debian_pkg():
  dpkg_src(
      name = "debian_stretch",
      arch = "amd64",
      distro = "stretch",
      sha256 = "9aea0e4c9ce210991c6edcb5370cb9b11e9e554a0f563e7754a4028a8fd0cb73",
      snapshot = "20171101T160520Z",
      url = "http://snapshot.debian.org/archive",
  )

  # download packages needed to create base docker image layers
  dpkg_list(
      name = "package_bundle",
      packages = [
          # libstdc++6 and dependencies (from https://packages.debian.org/stretch/libstdc%2B%2B6)
          # needed for admin CLI now for some reason
          "libstdc++6",
          "libgcc1",

          # iptables and dependencies (from https://packages.debian.org/stretch/iptables)
          "iptables",
          "libip4tc0",
          "libip6tc0",
          "libxtables12",
      ],
      sources = [
          "@debian_stretch//file:Packages.json",
      ],
  )

def _docker_dependencies_terraform():
  # download terraform binary to include in base docker image layer
  native.new_http_archive(
      name = "terraform_bin",
      url = "https://releases.hashicorp.com/terraform/0.10.8/terraform_0.10.8_linux_amd64.zip",
      sha256 = "b786c0cf936e24145fad632efd0fe48c831558cc9e43c071fffd93f35e3150db",
      build_file_content = """
filegroup(
    name = "bin",
    srcs = ["terraform"],
    visibility = ["//visibility:public"],
)
"""
  )
