git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    commit = "9556bc88d7d240d3bcbf07d24282953eb1687ef3",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")

go_rules_dependencies()
go_register_toolchains()

load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")
proto_register_toolchains()

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.3.0",
)

new_http_archive(
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

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
    container_repositories = "repositories",
)
container_repositories()

container_pull(
  name = "docker_git",
  registry = "index.docker.io",
  repository = "library/docker",
  tag = "17.06.2-ce-git"
)

container_pull(
  name = "debian_with_ssh",
  registry = "gcr.io/lattice-dev",
  repository = "debian-with-ssh",
  tag = "latest"
)

container_pull(
  name = "ubuntu_with_aws",
  registry = "gcr.io/lattice-dev",
  repository = "ubuntu-with-aws",
  tag = "latest"
)

container_pull(
  name = "ubuntu_with_iptables",
  registry = "gcr.io/lattice-dev",
  repository = "ubuntu-with-iptables",
  tag = "latest"
)

load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

go_repository(
    name = "com_github_mlab_lattice_core",
    commit = "e5687b550c2532a0695dae6cf45f1b1ff964a976",
    importpath = "github.com/mlab-lattice/core",
    remote = "git@github.com:mlab-lattice/core.git",
    vcs = "git",
)

go_repository(
    name = "io_k8s_apimachinery",
    # https://github.com/bazelbuild/rules_go/issues/964
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "9d38e20d609d27e00d4ec18f7b9db67105a2bde0",
    importpath = "k8s.io/apimachinery",
)

go_repository(
    name = "io_k8s_apiextensions_apiserver",
    # https://github.com/bazelbuild/rules_go/issues/964
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "79ecda8df91cd9304503d6f3e488341eabe2287f",
    importpath = "k8s.io/apiextensions-apiserver",
)

go_repository(
    name = "io_k8s_client_go",
    commit = "afb4606c45bae77c4dc2c15291d4d7d6d792196c",  # v5.0.0 branch HEAD
    importpath = "k8s.io/client-go",
)

go_repository(
    name = "io_k8s_api",
    # https://github.com/bazelbuild/rules_go/issues/964
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "fe29995db37613b9c5b2a647544cf627bfa8d299",  # Jul 19, 2017 (no releases)
    importpath = "k8s.io/api",
)

go_repository(
    name = "io_k8s_kube_openapi",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "868f2f29720b192240e18284659231b440f9cda5",
    importpath = "k8s.io/kube-openapi",
)

go_repository(
    name = "com_github_gin_gonic_gin",
    commit = "d459835d2b077e44f7c9b453505ee29881d5d12d",  # v1.2
    importpath = "github.com/gin-gonic/gin",
)

# Core dependencies
go_repository(
    name = "com_github_sergi_go_diff",
    commit = "feef008d51ad2b3778f85d387ccf91735543008d",
    importpath = "github.com/sergi/go-diff",
)

go_repository(
    name = "com_github_satori_go_uuid",
    commit = "5bf94b69c6b68ee1b541973bb8e1144db23a194b",
    importpath = "github.com/satori/go.uuid",
)

go_repository(
    name = "in_gopkg_src_d_go_git_v4",
    commit = "f9879dd043f84936a1f8acb8a53b74332a7ae135",
    importpath = "gopkg.in/src-d/go-git.v4",
)

go_repository(
    name = "org_golang_x_crypto",
    build_file_name = "BUILD.bazel",  # darwin build: case insensitive file system problem
    commit = "bd6f299fb381e4c3393d1c4b1f0b94f5e77650c8",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "com_github_xanzy_ssh_agent",
    commit = "ba9c9e33906f58169366275e3450db66139a31a9",
    importpath = "github.com/xanzy/ssh-agent",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    commit = "b8bc1bf767474819792c23f32d8286a45736f1c6",
    importpath = "github.com/mitchellh/go-homedir",
)

go_repository(
    name = "in_gopkg_src_d_go_billy_v3",
    commit = "c329b7bc7b9d24905d2bc1b85bfa29f7ae266314",
    importpath = "gopkg.in/src-d/go-billy.v3",
)

go_repository(
    name = "org_golang_x_text",
    build_file_name = "BUILD.bazel",  # darwin build: case insensitive file system problem
    commit = "88f656faf3f37f690df1a32515b479415e1a6769",
    importpath = "golang.org/x/text",
)

go_repository(
    name = "com_github_jbenet_go_context",
    commit = "d14ea06fba99483203c19d92cfcd13ebe73135f4",
    importpath = "github.com/jbenet/go-context",
)

go_repository(
    name = "org_golang_x_net",
    build_file_name = "BUILD.bazel",  # darwin build: case insensitive file system problem
    commit = "9dfe39835686865bff950a07b394c12a98ddc811",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "com_github_src_d_gcfg",
    commit = "f187355171c936ac84a82793659ebb4936bc1c23",
    importpath = "github.com/src-d/gcfg",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    commit = "ec4a0fea49c7b46c2aeb0b51aac55779c607e52b",
    importpath = "gopkg.in/warnings.v0",
)

go_repository(
    name = "com_github_docker_docker",
    commit = "f5ec1e2936dcbe7b5001c2b817188b095c700c27",
    importpath = "github.com/docker/docker",
)

# docker dependencies
# versions taken from https://github.com/moby/moby/blob/f5ec1e2936dcbe7b5001c2b817188b095c700c27/vendor.conf
go_repository(
    name = "com_github_docker_go_units",
    commit = "8a7beacffa3009a9ac66bad506b18ffdd110cf97",
    importpath = "github.com/docker/go-units",
)

go_repository(
    name = "com_github_docker_go_connections",
    commit = "ecb4cb2dd420ada7df7f2593d6c25441f65f69f2",
    importpath = "github.com/docker/go-connections",
)

go_repository(
    name = "com_github_docker_distribution",
    commit = "28602af35aceda2f8d571bad7ca37a54cf0250bc",
    importpath = "github.com/docker/distribution",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "839d9e913e063e28dfd0e6c7b7512793e0a48be9",
    importpath = "github.com/pkg/errors",
)

go_repository(
    name = "com_github_Sirupsen_logrus",
    tag = "v0.11.0",
    importpath = "github.com/Sirupsen/logrus",
)

# k8s dependencies
go_repository(
    name = "com_github_PuerkitoBio_purell",
    commit = "8a290539e2e8629dbc4e6bad948158f790ec31f4",
    importpath = "github.com/PuerkitoBio/purell",
)

go_repository(
    name = "com_github_PuerkitoBio_urlesc",
    commit = "5bd2802263f21d8788851d5305584c82a5c75d7e",
    importpath = "github.com/PuerkitoBio/urlesc",
)

go_repository(
    name = "com_github_emicklei_go_restful",
    commit = "ff4f55a206334ef123e4f79bbf348980da81ca46",
    importpath = "github.com/emicklei/go-restful",
)

go_repository(
    name = "com_github_go_openapi_jsonpointer",
    commit = "46af16f9f7b149af66e5d1bd010e3574dc06de98",
    importpath = "github.com/go-openapi/jsonpointer",
)

go_repository(
    name = "com_github_go_openapi_jsonreference",
    commit = "13c6e3589ad90f49bd3e3bbe2c2cb3d7a4142272",
    importpath = "github.com/go-openapi/jsonreference",
)

go_repository(
    name = "com_github_go_openapi_spec",
    commit = "6aced65f8501fe1217321abf0749d354824ba2ff",
    importpath = "github.com/go-openapi/spec",
)

go_repository(
    name = "com_github_go_openapi_swag",
    commit = "1d0bd113de87027671077d3c71eb3ac5d7dbba72",
    importpath = "github.com/go-openapi/swag",
)

go_repository(
    name = "com_github_gogo_protobuf",
    commit = "c0656edd0d9eab7c66d1eb0c568f9039345796f7",
    importpath = "github.com/gogo/protobuf",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "44145f04b68cf362d9c4df2182967c2275eaefed",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "17ce1425424ab154092bbb43af630bd647f3bb0d",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "com_github_google_gofuzz",
    commit = "44d81051d367757e1c7c6a5a86423ece9afcf63c",
    importpath = "github.com/google/gofuzz",
)

go_repository(
    name = "com_github_mailru_easyjson",
    commit = "d5b7844b561a7bc640052f1b935f7b800330d7e0",
    importpath = "github.com/mailru/easyjson",
)

go_repository(
    name = "org_golang_x_net",
    commit = "f2499483f923065a842d38eb4c7f1927e6fc6e6d",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_text",
    build_file_name = "BUILD.bazel",  # darwin build: case insensitive file system problem
    commit = "2910a502d2bf9e43193af9d68ca516529614eed3",
    importpath = "golang.org/x/text",
)

go_repository(
    name = "in_gopkg_inf_v0",
    commit = "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
    importpath = "gopkg.in/inf.v0",
)

go_repository(
    name = "com_github_spf13_cobra",
    commit = "2df9a531813370438a4d79bfc33e21f58063ed87",
    importpath = "github.com/spf13/cobra",
)

go_repository(
    name = "com_github_spf13_pflag",
    commit = "e57e3eeb33f795204c1ca35f56c44f83227c6e66",
    importpath = "github.com/spf13/pflag",
)

go_repository(
    name = "com_github_juju_ratelimit",
    commit = "5b9ff866471762aa2ab2dced63c9fb6f53921342",  # May 23, 2017 (no releases)
    importpath = "github.com/juju/ratelimit",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "0a025b7e63adc15a622f29b0b2c4c3848243bbf6",  # Aug 13, 2016 (no releases)
    importpath = "github.com/hashicorp/golang-lru",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "346938d642f2ec3594ed81d874461961cd0faa76",  # Nov 14, 2016 (1.1.0)
    importpath = "github.com/davecgh/go-spew",
)

go_repository(
    name = "com_github_ugorji_go",
    commit = "708a42d246822952f38190a8d8c4e6b16a0e600c",  # Mar 12, 2017 (no releases)
    importpath = "github.com/ugorji/go",
)

go_repository(
    name = "com_github_ghodss_yaml",
    commit = "04f313413ffd65ce25f2541bfd2b2ceec5c0908c",  # Dec 6, 2016 (no releases)
    importpath = "github.com/ghodss/yaml",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "14227de293ca979cf205cd88769fe71ed96a97e2",  # Jan 24, 2017 (no releases)
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "com_github_googleapis_gnostic",
    # https://github.com/bazelbuild/rules_go/issues/964
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "ee43cbb60db7bd22502942cccbc39059117352ab",
    importpath = "github.com/googleapis/gnostic",
)

go_repository(
    name = "com_github_emicklei_go_restful_swagger12",
    commit = "dcef7f55730566d41eae5db10e7d6981829720f6",
    importpath = "github.com/emicklei/go-restful-swagger12",
)

go_repository(
    name = "com_github_howeyc_gopass",
    commit = "bf9dde6d0d2c004a008c27aaee91170c786f6db8",
    importpath = "github.com/howeyc/gopass",
)

go_repository(
    name = "com_github_imdario_mergo",
    commit = "e3000cb3d28c72b837601cac94debd91032d19fe",
    importpath = "github.com/imdario/mergo",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "81e90905daefcd6fd217b62423c0908922eadb30",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "9aade4d3a3b7e6d876cd3823ad20ec45fc035402",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "com_github_pborman_uuid",
    commit = "e790cca94e6cc75c7064b1332e63811d4aae1a53",
    importpath = "github.com/pborman/uuid",
)

go_repository(
    name = "com_github_sergi_go_diff",
    commit = "feef008d51ad2b3778f85d387ccf91735543008d",
    importpath = "github.com/sergi/go-diff",
)

go_repository(
    name = "com_github_peterbourgon_diskv",
    commit = "5f041e8faa004a95c88a202771f4cc3e991971e6",
    importpath = "github.com/peterbourgon/diskv",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "787624de3eb7bd915c329cba748687a3b22666a6",
    importpath = "github.com/gregjones/httpcache",
)

go_repository(
    name = "com_github_google_btree",
    commit = "7d79101e329e5a3adf994758c578dab82b90c017",
    importpath = "github.com/google/btree",
)

go_repository(
    name = "com_github_json_iterator_go",
    commit = "36b14963da70d11297d313183d7e6388c8510e1e",
    importpath = "github.com/json-iterator/go",
)

# Gin dependencies. Commits from v1.2 vendor/vendor.json
go_repository(
    name = "com_github_mattn_go_isatty",
    commit = "57fdcb988a5c543893cc61bce354a6e24ab70022",
    importpath = "github.com/mattn/go-isatty",
)

go_repository(
    name = "com_github_gin_contrib_sse",
    commit = "22d885f9ecc78bf4ee5d72b937e4bbcdc58e8cae",
    importpath = "github.com/gin-contrib/sse",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "130e6b02ab059e7b717a096f397c5b60111cae74",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "a5b47d31c556af34a302ce5d659e6fea44d90de0",
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "in_gopkg_go_playground_validator_v8",
    commit = "5f57d2222ad794d0dffb07e664ea05e2ee07d60c",
    importpath = "gopkg.in/go-playground/validator.v8",
)

go_repository(
    name = "com_github_ugorji_go",
    commit = "c88ee250d0221a57af388746f5cf03768c21d6e2",
    importpath = "github.com/ugorji/go",
)
