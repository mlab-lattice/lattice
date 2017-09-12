git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.1.0",
)

load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_repositories",
    "docker_pull",
)

docker_repositories()

docker_pull(
    name = "official_ubuntu",
    registry = "index.docker.io",
    repository = "library/ubuntu",
    tag = "14.04",
)

git_repository(
    name = "io_bazel_rules_go",
    #    tag = "0.5.3",
    commit = "abfa5e953f870fd45618bc1d30ae971f37f5a34d",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_rules_dependencies",
    "go_register_toolchains",
    "go_repository",
)

go_rules_dependencies()

go_register_toolchains()

go_repository(
    name = "com_github_mlab_lattice_core",
    commit = "c005ec1b4814af658fa8488f8d883913fb57bfd2",
    importpath = "github.com/mlab-lattice/core",
    remote = "git@github.com:mlab-lattice/core.git",
    vcs = "git",
)

go_repository(
    name = "io_k8s_apimachinery",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    commit = "dc1f89aff9a7509782bde3b68824c8043a3e58cc",
    importpath = "k8s.io/apimachinery",
)

go_repository(
    name = "io_k8s_apiextensions_apiserver",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    commit = "79ecda8df91cd9304503d6f3e488341eabe2287f",
    importpath = "k8s.io/apiextensions-apiserver",
)

go_repository(
    name = "io_k8s_client_go",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    commit = "7c69e980210777a6292351ac6873de083526f08e",  # Jul 18, 2017 (no releases)
    importpath = "k8s.io/client-go",
)

go_repository(
    name = "io_k8s_api",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    commit = "4d5cc6efc5e84aa19fb1bd3f911c16a6723c1bb7",  # Jul 19, 2017 (no releases)
    importpath = "k8s.io/api",
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
