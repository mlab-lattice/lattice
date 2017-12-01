# go (needs to be called first since some of the docker commands use it)
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


# docker
git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    commit = "caa55f1e00e5909dbd689f298e2c6d3ef3e65d81",
)
load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)
_go_image_repos()


# distroless repo contains rules about downloading debs
git_repository(
    name = "distroless",
    remote = "https://github.com/GoogleCloudPlatform/distroless.git",
    commit = "e5854b38a12bb37adaf0edb193f97b32a3bcaee0",
)
load(
    "@distroless//package_manager:package_manager.bzl",
    "package_manager_repositories",
    "dpkg_src",
    "dpkg_list",
)
package_manager_repositories()

# this will download debs from a snapshot of the debian archive
# more information: https://github.com/GoogleCloudPlatform/distroless/tree/master/package_manager
dpkg_src(
    name = "debian_stretch",
    arch = "amd64",
    distro = "stretch",
    sha256 = "9aea0e4c9ce210991c6edcb5370cb9b11e9e554a0f563e7754a4028a8fd0cb73",
    snapshot = "20171101T160520Z",
    url = "http://snapshot.debian.org/archive",
)

# download actual debs needed to create base docker image layers
dpkg_list(
    name = "package_bundle",
    packages = [
        # iptables and dependencies (from https://packages.debian.org/sid/iptables)
        "iptables",
        "libip4tc0",
        "libip6tc0",
        "libxtables12",
    ],
    sources = [
        "@debian_stretch//file:Packages.json",
    ],
)


# download terraform binary to include in base docker image layer
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


# go dependencies

# direct github.com/mlab-lattice/system dependencies
go_repository(
    name = "com_github_mlab_lattice_core",
    commit = "fe944f3e269aae519c0b339d6a97e1e2951dd313",
    importpath = "github.com/mlab-lattice/core",
    remote = "git@github.com:mlab-lattice/core.git",
    vcs = "git",
)

# also depended upon by k8s.io
# jumping ahead of their requirement to include: https://github.com/spf13/cobra/pull/502
go_repository(
    name = "com_github_spf13_cobra",
    commit = "1be1d2841c773c01bee8289f55f7463b6e2c2539",
    importpath = "github.com/spf13/cobra",
)

# update this when you update cobra
go_repository(
    name = "com_github_spf13_pflag",
    commit = "4c012f6dcd9546820e378d0bdda4d8fc772cdfea",
    importpath = "github.com/spf13/pflag",
)

go_repository(
    name = "com_github_satori_go_uuid",
    commit = "5bf94b69c6b68ee1b541973bb8e1144db23a194b",
    importpath = "github.com/satori/go.uuid",
)

go_repository(
    name = "com_github_docker_docker",
    tag = "v17.03.2-ce",
    importpath = "github.com/docker/docker",
)

go_repository(
    name = "com_github_gin_gonic_gin",
    tag = "v1.2",
    importpath = "github.com/gin-gonic/gin",
)

go_repository(
    name = "com_github_fatih_color",
    tag = "v1.5.0",
    importpath = "github.com/fatih/color",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    tag = "v1.12.35",
    importpath = "github.com/aws/aws-sdk-go",
)

go_repository(
    name = "com_github_coreos_go_iptables",
    # repo has a file named "build" so have to force gazelle to generate a BUILD.bazel file
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    commit = "17b936e6ccb6f6e424f7d89c614164e796df1661",
    importpath = "github.com/coreos/go-iptables",
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


# github.com/mlab-lattice/core dependencies
go_repository(
    name = "in_gopkg_src_d_go_git_v4",
    commit = "f9879dd043f84936a1f8acb8a53b74332a7ae135",
    importpath = "gopkg.in/src-d/go-git.v4",
)

go_repository(
    name = "com_github_sergi_go_diff",
    commit = "feef008d51ad2b3778f85d387ccf91735543008d",
    importpath = "github.com/sergi/go-diff",
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
    name = "com_github_jbenet_go_context",
    commit = "d14ea06fba99483203c19d92cfcd13ebe73135f4",
    importpath = "github.com/jbenet/go-context",
)

# also depended on by k8s.io, version taken from same place as the rest of k8s.io dependencies
go_repository(
    name = "org_golang_x_crypto",
    commit = "81e90905daefcd6fd217b62423c0908922eadb30",
    importpath = "golang.org/x/crypto",
)

# also depended on by k8s.io, version taken from same place as the rest of k8s.io dependencies
go_repository(
    name = "org_golang_x_text",
    commit = "b19bf474d317b857955b12035d2c5acb57ce8b01",
    importpath = "golang.org/x/text",
)

# also depended on by k8s.io, version taken from same place as the rest of k8s.io dependencies
go_repository(
    name = "org_golang_x_net",
    commit = "1c05540f6879653db88113bc4a2b70aec4bd491f",
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


# github.com/docker/docker dependencies
# commits taken from https://github.com/docker/docker/blob/v17.03.2-ce/vendor.conf
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

# Commit in vendor.conf is 54296cf40ad8143b62dbcaa1d90e520a2136ddfe, but bazel did not like this (said it was not a tree).
# Seems to be cherry pick of c7ebda72acad31929e35b4fc6c2739013cf4fadd, so using that instead.
go_repository(
    name = "com_github_opencontainers_runc",
    commit = "c7ebda72acad31929e35b4fc6c2739013cf4fadd",
    importpath = "github.com/opencontainers/runc",
)


# k8s.io dependencies
# commits from https://github.com/kubernetes/kubernetes/blob/9befc2b8928a9426501d3bf62f72849d5cbcd5a3/Godeps/Godeps.json
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
    commit = "4bd1920723d7b7c925de087aa32e2187708897f7",
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
    name = "in_gopkg_inf_v0",
    commit = "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
    importpath = "gopkg.in/inf.v0",
)

go_repository(
    name = "com_github_juju_ratelimit",
    commit = "5b9ff866471762aa2ab2dced63c9fb6f53921342",
    importpath = "github.com/juju/ratelimit",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4",
    importpath = "github.com/hashicorp/golang-lru",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "782f4967f2dc4564575ca782fe2d04090b5faca8",
    importpath = "github.com/davecgh/go-spew",
)

go_repository(
    name = "com_github_ugorji_go",
    commit = "ded73eae5db7e7a0ef6f55aace87a2873c5d2b74",
    importpath = "github.com/ugorji/go",
)

go_repository(
    name = "com_github_ghodss_yaml",
    commit = "73d445a93680fa1a78ae23a5839bad48f32ba1ee",
    importpath = "github.com/ghodss/yaml",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "53feefa2559fb8dfa8d81baad31be332c97d6c77",
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "com_github_googleapis_gnostic",
    # https://github.com/bazelbuild/rules_go/issues/964
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable",
    commit = "0c5108395e2debce0d731cf0287ddf7242066aba",
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
    commit = "6633656539c1639d9d78127b7d47c622b5d7b6dc",
    importpath = "github.com/imdario/mergo",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "7ddbeae9ae08c6a06a59597f0c9edbc5ff2444ce",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "com_github_pborman_uuid",
    commit = "ca53cad383cad2479bbba7f7a1a05797ec1386e4",
    importpath = "github.com/pborman/uuid",
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


# Gin dependencies.
# commits from https://github.com/gin-gonic/gin/blob/d459835d2b077e44f7c9b453505ee29881d5d12d/vendor/vendor.json
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


# github.com/fatih/go dependencies
# commits taken from: https://github.com/gin-gonic/gin/blob/v1.2/vendor/vendor.json
go_repository(
    name = "com_github_mattn_go_colorable",
    commit = "5411d3eea5978e6cdc258b30de592b60df6aba96",
    importpath = "github.com/mattn/go-colorable",
)


# github.com/aws/aws-sdk-go dependencies
# commits taken from: https://github.com/aws/aws-sdk-go/blob/v1.12.35/Gopkg.lock
go_repository(
    name = "com_github_go_ini_ini",
    commit = "300e940a926eb277d3901b20bdfcc54928ad3642",
    importpath = "github.com/go-ini/ini",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74",
    importpath = "github.com/jmespath/go-jmespath",
)
