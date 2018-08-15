load("@bazel_gazelle//:deps.bzl", "go_repository")

GO_DEPENDENCIES = {
    "overlaps": {
        # github.com/envoyproxy/go-control-plane:v0.2 -> 342cbe0a04158f6dcb03ca0079991a51a4248c02
        # k8s.io:v1.9.3 -> c0656edd0d9eab7c66d1eb0c568f9039345796f7
        "github.com/gogo/protobuf": {
            "name": "com_github_gogo_protobuf",
            "commit": "342cbe0a04158f6dcb03ca0079991a51a4248c02",
            "importpath": "github.com/gogo/protobuf",
            "build_file_proto_mode": "disable",
        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> ab9f9a6dab164b7d1246e0e688b0ab7b94d8553e
        # k8s.io:v1.9.3 -> 1643683e1b54a9e88ad26d98f81400c8c9d9f4f9
#        "github.com/golang/protobuf": {
#            "name": "com_github_golang_protobuf",
#            "commit": "b4deda0973fb4c70b50d226b1af49f3da59f5265",
#            "importpath": "github.com/golang/protobuf",
#        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> c155da19408a8799da419ed3eeb0cb5db0ad5dbc
        # github.com/docker/docker:1a57535aa277e0f2a3c1922c736551148c5b4351 -> v1.0.3
        "github.com/sirupsen/logrus": {
            "name": "com_github_sirupsen_logrus",
            "commit": "c155da19408a8799da419ed3eeb0cb5db0ad5dbc",
            "importpath": "github.com/sirupsen/logrus",
        },
        # github.com/mlab-lattice/lattice -> 81e90905daefcd6fd217b62423c0908922eadb30
        # github.com/envoyproxy/go-control-plane:v0.2 -> 5119cf507ed5294cc409c092980c7497ee5d6fd2
        "golang.org/x/crypto": {
            "name": "org_golang_x_crypto",
            "commit": "81e90905daefcd6fd217b62423c0908922eadb30",
            "importpath": "golang.org/x/crypto",
        },
        # github.com/docker/docker:1a57535aa277e0f2a3c1922c736551148c5b4351 -> 5561cd9b4330353950f399814f427425c0a26fd2
        # k8s.io:v1.9.3 -> ?
        # github.com/envoyproxy/go-control-plane:v0.2 -> 5ccada7d0a7ba9aeb5d3aca8d3501b4c2a509fec
        "golang.org/x/net": {
            "name": "org_golang_x_net",
            "commit": "5561cd9b4330353950f399814f427425c0a26fd2",
            "importpath": "golang.org/x/net",
        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> 37707fdb30a5b38865cfb95e5aab41707daec7fd
        # k8s.io:v1.9.3 -> 95c6576299259db960f6c5b9b69ea52422860fce
        "golang.org/x/sys": {
            "name": "org_golang_x_sys",
            "commit": "37707fdb30a5b38865cfb95e5aab41707daec7fd",
            "importpath": "golang.org/x/sys",
        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> e19ae1496984b1c655b8044a65c0300a3c878dd3
        # k8s.io:v1.9.3 -> b19bf474d317b857955b12035d2c5acb57ce8b01
        "golang.org/x/text": {
            "name": "org_golang_x_text",
            "commit": "e19ae1496984b1c655b8044a65c0300a3c878dd3",
            "importpath": "golang.org/x/text",
        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> a8101f21cf983e773d0c1133ebc5424792003214
        # k8s.io:v1.9.3 -> 09f6ed296fc66555a25fe4ce95173148778dfa85
        "google.golang.org/genproto": {
            "name": "org_golang_google_genproto",
            "commit": "a8101f21cf983e773d0c1133ebc5424792003214",
            "importpath": "google.golang.org/genproto",
        },
        # github.com/envoyproxy/go-control-plane:v0.2 -> 1e2570b1b19ade82d8dbb31bba4e65e9f9ef5b34
        # k8s.io:v1.9.3 -> 5b3c4e850e90a4cf6a20ebd46c8b32a0a3afcb9e
        "google.golang.org/grpc": {
            "name": "org_golang_google_grpc",
            "commit": "1e2570b1b19ade82d8dbb31bba4e65e9f9ef5b34",
            "importpath": "google.golang.org/grpc",
        },
    },
    "github.com/mlab-lattice/lattice": {
        "github.com/aws/aws-sdk-go": {
            "name": "com_github_aws_aws_sdk_go",
            "tag": "v1.12.35",
            "importpath": "github.com/aws/aws-sdk-go",
        },
        "github.com/blang/semver": {
            "name": "com_github_blang_semver",
            "tag": "v3.5.1",
            "importpath": "github.com/blang/semver",
        },
        "github.com/briandowns/spinner": {
            "name": "com_github_briandowns_spinner",
            "commit": "b3ed21e9d3b2af92b077f7d21edd80599b8fb97f",
            "importpath": "github.com/briandowns/spinner",
        },
        # master HEAD as of 3/6/18
        "github.com/buger/goterm": {
            "name": "com_github_buger_goterm",
            "commit": "2f70ee476a85f692e8d4265963c615d8a57652da",
            "importpath": "github.com/buger/goterm",
        },
        "github.com/coreos/go-iptables": {
           "name": "com_github_coreos_go_iptables",
           # repo has a file named "build" so have to force gazelle to generate a BUILD.bazel file
           "build_file_generation": "on",
           "build_file_name": "BUILD.bazel",
           "commit": "17b936e6ccb6f6e424f7d89c614164e796df1661",
           "importpath": "github.com/coreos/go-iptables",
        },
        "github.com/deckarep/golang-set": {
            "name": "com_github_deckarep_golang_set",
            "commit": "1d4478f51bed434f1dadf96dcd9b43aabac66795",
            "importpath": "github.com/deckarep/golang-set",
        },
        "github.com/docker/docker": {
            "name": "com_github_docker_docker",
            # corresponds to docker v18.05.0-ce
            "commit": "1a57535aa277e0f2a3c1922c736551148c5b4351",
            "importpath": "github.com/docker/docker",
        },
        "github.com/envoyproxy/go-control-plane": {
            "name": "com_github_envoyproxy_go_control_plane",
            "tag": "v0.2",
            "importpath": "github.com/envoyproxy/go-control-plane",
        },
        "github.com/fatih/color": {
            "name": "com_github_fatih_color",
            "tag": "v1.5.0",
            "importpath": "github.com/fatih/color",
        },
        "github.com/gin-gonic/gin": {
            "name": "com_github_gin_gonic_gin",
            "tag": "v1.2",
            "importpath": "github.com/gin-gonic/gin",
        },
        "github.com/satori/go.uuid": {
            "name": "com_github_satori_go_uuid",
            "commit": "5bf94b69c6b68ee1b541973bb8e1144db23a194b",
            "importpath": "github.com/satori/go.uuid",
        },
        "github.com/sergi/go-diff": {
            "name": "com_github_sergi_go_diff",
            "commit": "feef008d51ad2b3778f85d387ccf91735543008d",
            "importpath": "github.com/sergi/go-diff",
        },
        # also depended upon by k8s.io
        # jumping ahead of their requirement to include: https://github.com/spf13/cobra/pull/502
        "github.com/spf13/cobra": {
            "name": "com_github_spf13_cobra",
            "commit": "1be1d2841c773c01bee8289f55f7463b6e2c2539",
            "importpath": "github.com/spf13/cobra",
        },
        "github.com/tfogo/tablewriter": {
            "name": "com_github_tfogo_tablewriter",
            "commit": "4776fb554dc2ca114fbce4738142a47de1ea0929",
            "importpath": "github.com/tfogo/tablewriter",
        },
        "github.com/tidwall/gjson": {
            "name": "com_github_tidwall_gjson",
            "tag": "v1.0.6",
            "importpath": "github.com/tidwall/gjson",
        },
        "gopkg.in/src-d/go-git.v4": {
            "name": "in_gopkg_src_d_go_git_v4",
            "commit": "f9879dd043f84936a1f8acb8a53b74332a7ae135",
            "importpath": "gopkg.in/src-d/go-git.v4",
        },
        "k8s.io/api": {
            "name": "io_k8s_api",
            # https://github.com/bazelbuild/rules_go/issues/964
            "build_file_proto_mode": "disable",
            "tag": "kubernetes-1.11.0",
            "importpath": "k8s.io/api",
        },
        "k8s.io/apimachinery": {
            "name": "io_k8s_apimachinery",
            # https://github.com/bazelbuild/rules_go/issues/964
            "build_file_proto_mode": "disable",
            "tag": "kubernetes-1.11.0",
            "importpath": "k8s.io/apimachinery",
        },
        "k8s.io/apiextensions-apiserver": {
            "name": "io_k8s_apiextensions_apiserver",
            "build_external": "external",
            # https://github.com/bazelbuild/rules_go/issues/964
            "build_file_proto_mode": "disable",
            "tag": "kubernetes-1.11.0",
            "importpath": "k8s.io/apiextensions-apiserver",
        },
        "k8s.io/client-go": {
            "name": "io_k8s_client_go",
            "tag": "kubernetes-1.11.0",
            "importpath": "k8s.io/client-go",
        },
        "k8s.io/kubernetes": {
            "name": "io_k8s_kubernetes",
            "build_file_generation": "on",
            "build_file_name": "BUILD.bazel",
            "build_external": "external",
            "tag": "v1.11.0",
            "importpath": "k8s.io/kubernetes",
        },

        # testing dependencies
        "github.com/onsi/ginkgo": {
            "name": "com_github_onsi_ginkgo",
            "tag": "v1.4.0",
            "importpath": "github.com/onsi/ginkgo",
        },
        "github.com/onsi/gomega": {
            "name": "com_github_onsi_gomega",
            "tag": "v1.3.0",
            "importpath": "github.com/onsi/gomega",
        },
    },

    # commits taken from: https://github.com/aws/aws-sdk-go/blob/v1.12.35/Gopkg.lock
    "github.com/aws/aws-sdk-go": {
        "github.com/go-ini/ini": {
            "name": "com_github_go_ini_ini",
            "commit": "300e940a926eb277d3901b20bdfcc54928ad3642",
            "importpath": "github.com/go-ini/ini",
        },
        "github.com/jmespath/go-jmespath": {
            "name": "com_github_jmespath_go_jmespath",
            "commit": "0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74",
            "importpath": "github.com/jmespath/go-jmespath",
        },
    },

    # no dependency versions listed, taken from master HEAD
    "github.com/deckarep/golang-set": {
        "github.com/golang/groupcache": {
            "name": "com_github_golang_groupcache",
            "commit": "84a468cf14b4376def5d68c722b139b881c450a4",
            "importpath": "github.com/golang/groupcache",
        },
    },

    # Getting the right commits is a little tricky. First go to https://github.com/docker/docker-ce and find
    # the corresponding release. Then go to components/engine and look at the most recent commit.
    # At the bottom of the commit message it should say what commit is was cherry picked from.
    # Go to github.com/moby/moby and go to that commit and check vendor.conf
    # Taken from https://github.com/docker/docker-ce/tree/v18.05.0-ce/components/engine
    #   -> https://github.com/docker/docker-ce/commit/c1ef5d203fff0b7c86c403d614a1e81a014fae56
    #   -> https://github.com/moby/moby/blob/1a57535aa277e0f2a3c1922c736551148c5b4351/vendor.conf
    "github.com/docker/docker": {
        "github.com/docker/distribution": {
            "name": "com_github_docker_distribution",
            "commit": "edc3ab29cdff8694dd6feb85cfeb4b5f1b38ed9c",
            "importpath": "github.com/docker/distribution",
        },
        "github.com/docker/go-connections": {
            "name": "com_github_docker_go_connections",
            "commit": "7beb39f0b969b075d1325fecb092faf27fd357b6",
            "importpath": "github.com/docker/go-connections",
        },
        "github.com/docker/go-units": {
            "name": "com_github_docker_go_units",
            "commit": "9e638d38cf6977a37a8ea0078f3ee75a7cdb2dd1",
            "importpath": "github.com/docker/go-units",
        },
        "github.com/opencontainers/runc": {
            "name": "com_github_opencontainers_runc",
            "commit": "4fc53a81fb7c994640722ac585fa9ca548971871",
            "importpath": "github.com/opencontainers/runc",
        },
        "github.com/pkg/errors": {
            "name": "com_github_pkg_errors",
            "commit": "839d9e913e063e28dfd0e6c7b7512793e0a48be9",
            "importpath": "github.com/pkg/errors",
        },
        "github.com/opencontainers/go-digest": {
            "name": "com_github_opencontainers_go_digest",
            "tag": "v1.0.0-rc1",
            "importpath": "github.com/opencontainers/go-digest",
        },
        "github.com/Nvveen/Gotty": {
            "name": "com_github_Nvveen_Gotty",
            "commit": "a8b993ba6abdb0e0c12b0125c603323a71c7790c",
            "importpath": "github.com/Nvveen/Gotty",
            # looks like github.com/Nvveen/Gotty is what is being included:
            # https://github.com/moby/moby/blob/6e7715d65ba892a47d355e16bf9ad87fb537a2d0/pkg/jsonmessage/jsonmessage.go#L11
            # but that it is being aliased in vendor.conf:
            # https://github.com/moby/moby/blob/6e7715d65ba892a47d355e16bf9ad87fb537a2d0/vendor.conf#L142
            "vcs": "git",
            "remote": "https://github.com/ijc25/Gotty",
        },
        "github.com/docker/libtrust": {
            "name": "com_github_docker_libtrust",
            "commit": "9cbd2a1374f46905c68a4eb3694a130610adc62a",
            "importpath": "github.com/docker/libtrust",
        },
    },

    # commits taken from: https://github.com/envoyproxy/go-control-plane/blob/v0.2/glide.lock
    "github.com/envoyproxy/go-control-plane": {
        "github.com/envoyproxy/data-plane-api": {
            "name": "com_github_envoyproxy_data_plane_api",
            "commit": "22d98ce77228bde6984dfee38a285f501e748343",
            "importpath": "github.com/envoyproxy/data-plane-api",
        },
        "github.com/gogo/googleapis": {
            "name": "com_github_gogo_googleapis",
            "commit": "0cd9801be74a10d5ac39d69626eac8255ffcd502",
            "importpath": "github.com/gogo/googleapis",
            "build_file_proto_mode": "disable",
        },
        "github.com/lyft/protoc-gen-validate": {
            "name": "com_github_lyft_protoc_gen_validate",
            "commit": "2463485ae0c04eea7977c7f521549229f659e39a",
            "importpath": "github.com/lyft/protoc-gen-validate",
            "build_file_proto_mode": "disable",
        },
        "istio.io/gogo-genproto": {
            "name": "io_istio_gogo_genproto",
            "commit": "dd2b5b6a8f1400838165245e704a011a30154a44",
            "importpath": "istio.io/gogo-genproto",
        },
    },

    # commits taken from: https://github.com/gin-gonic/gin/blob/v1.2/vendor/vendor.json
    "github.com/fatih/color": {
        "github.com/mattn/go-colorable": {
            "name": "com_github_mattn_go_colorable",
            "commit": "5411d3eea5978e6cdc258b30de592b60df6aba96",
            "importpath": "github.com/mattn/go-colorable",
        },
    },

    # commits from https://github.com/gin-gonic/gin/blob/d459835d2b077e44f7c9b453505ee29881d5d12d/vendor/vendor.json
    "github.com/gin-gonic/gin": {
        # also depended upon by github.com/fatih/color
        "github.com/gin-contrib/sse": {
            "name": "com_github_gin_contrib_sse",
            "commit": "22d885f9ecc78bf4ee5d72b937e4bbcdc58e8cae",
            "importpath": "github.com/gin-contrib/sse",
        },
        "github.com/mattn/go-isatty": {
            "name": "com_github_mattn_go_isatty",
            "commit": "57fdcb988a5c543893cc61bce354a6e24ab70022",
            "importpath": "github.com/mattn/go-isatty",
        },
        "gopkg.in/go-playground/validator.v8": {
            "name": "in_gopkg_go_playground_validator_v8",
            "commit": "5f57d2222ad794d0dffb07e664ea05e2ee07d60c",
            "importpath": "gopkg.in/go-playground/validator.v8",
        },
    },

    # no dependency versions listed, taken from master HEAD
    "github.com/olekukonko/tablewriter": {
        "github.com/mattn/go-runewidth": {
            "name": "com_github_mattn_go_runewidth",
            "commit": "97311d9f7767e3d6f422ea06661bc2c7a19e8a5d",
            "importpath": "github.com/mattn/go-runewidth",
        },
    },

    "github.com/spf13/cobra": {
        "github.com/spf13/pflag": {
            "name": "com_github_spf13_pflag",
            "commit": "4c012f6dcd9546820e378d0bdda4d8fc772cdfea",
            "importpath": "github.com/spf13/pflag",
        },
    },

    "github.com/tidwall/gjson": {
        "github.com/tidwall/match": {
            "name": "com_github_tidwall_match",
            "commit": "1731857f09b1f38450e2c12409748407822dc6be",
            "importpath": "github.com/tidwall/match",
        },
    },

    # could not find a list of dependency versions, so mostly took master HEAD
    "gopkg.in/src-d/go-git.v4": {
        "github.com/jbenet/go-context": {
            "name": "com_github_jbenet_go_context",
            "commit": "d14ea06fba99483203c19d92cfcd13ebe73135f4",
            "importpath": "github.com/jbenet/go-context",
        },
        "github.com/mitchellh/go-homedir": {
            "name": "com_github_mitchellh_go_homedir",
            "commit": "b8bc1bf767474819792c23f32d8286a45736f1c6",
            "importpath": "github.com/mitchellh/go-homedir",
        },
        "github.com/src-d/gcfg": {
            "name": "com_github_src_d_gcfg",
            "commit": "f187355171c936ac84a82793659ebb4936bc1c23",
            "importpath": "github.com/src-d/gcfg",
        },
        "github.com/xanzy/ssh-agent": {
            "name": "com_github_xanzy_ssh_agent",
            "commit": "ba9c9e33906f58169366275e3450db66139a31a9",
            "importpath": "github.com/xanzy/ssh-agent",
        },
        "gopkg.in/src-d/go-billy.v3": {
            "name": "in_gopkg_src_d_go_billy_v3",
            "commit": "c329b7bc7b9d24905d2bc1b85bfa29f7ae266314",
            "importpath": "gopkg.in/src-d/go-billy.v3",
        },
        "gopkg.in/warnings.v0": {
            "name": "in_gopkg_warnings_v0",
            "commit": "ec4a0fea49c7b46c2aeb0b51aac55779c607e52b",
            "importpath": "gopkg.in/warnings.v0",
        },
    },

    # commits taken from https://github.com/kubernetes/kubernetes/blob/v1.9.3/Godeps/Godeps.json
    "k8s.io": {
        "github.com/davecgh/go-spew": {
            "name": "com_github_davecgh_go_spew",
            "commit": "782f4967f2dc4564575ca782fe2d04090b5faca8",
            "importpath": "github.com/davecgh/go-spew",
        },
        "github.com/emicklei/go-restful": {
            "name": "com_github_emicklei_go_restful",
            "commit": "ff4f55a206334ef123e4f79bbf348980da81ca46",
            "importpath": "github.com/emicklei/go-restful",
        },
        "github.com/emicklei/go-restful-swagger12": {
            "name": "com_github_emicklei_go_restful_swagger12",
            "commit": "dcef7f55730566d41eae5db10e7d6981829720f6",
            "importpath": "github.com/emicklei/go-restful-swagger12",
        },
        "github.com/ghodss/yaml": {
            "name": "com_github_ghodss_yaml",
            "commit": "73d445a93680fa1a78ae23a5839bad48f32ba1ee",
            "importpath": "github.com/ghodss/yaml",
        },
        "github.com/go-openapi/jsonpointer": {
            "name": "com_github_go_openapi_jsonpointer",
            "commit": "46af16f9f7b149af66e5d1bd010e3574dc06de98",
            "importpath": "github.com/go-openapi/jsonpointer",
        },
        "github.com/go-openapi/jsonreference": {
            "name": "com_github_go_openapi_jsonreference",
            "commit": "13c6e3589ad90f49bd3e3bbe2c2cb3d7a4142272",
            "importpath": "github.com/go-openapi/jsonreference",
        },
        "github.com/go-openapi/spec": {
            "name": "com_github_go_openapi_spec",
            "commit": "1de3e0542de65ad8d75452a595886fdd0befb363",
            "importpath": "github.com/go-openapi/spec",
        },
        "github.com/go-openapi/swag": {
            "name": "com_github_go_openapi_swag",
            "commit": "f3f9494671f93fcff853e3c6e9e948b3eb71e590",
            "importpath": "github.com/go-openapi/swag",
        },
        "github.com/google/btree": {
            "name": "com_github_google_btree",
            "commit": "7d79101e329e5a3adf994758c578dab82b90c017",
            "importpath": "github.com/google/btree",
        },
        "github.com/google/gofuzz": {
            "name": "com_github_google_gofuzz",
            "commit": "44d81051d367757e1c7c6a5a86423ece9afcf63c",
            "importpath": "github.com/google/gofuzz",
        },
        "github.com/googleapis/gnostic": {
            "name": "com_github_googleapis_gnostic",
            # https://github.com/bazelbuild/rules_go/issues/964
            "build_file_proto_mode": "disable",
            "commit": "0c5108395e2debce0d731cf0287ddf7242066aba",
            "importpath": "github.com/googleapis/gnostic",
        },
        "github.com/gregjones/httpcache": {
            "name": "com_github_gregjones_httpcache",
            "commit": "787624de3eb7bd915c329cba748687a3b22666a6",
            "importpath": "github.com/gregjones/httpcache",
        },
        "github.com/hashicorp/golang-lru": {
            "name": "com_github_hashicorp_golang_lru",
            "commit": "a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4",
            "importpath": "github.com/hashicorp/golang-lru",
        },
        "github.com/imdario/mergo": {
            "name": "com_github_imdario_mergo",
            "commit": "6633656539c1639d9d78127b7d47c622b5d7b6dc",
            "importpath": "github.com/imdario/mergo",
        },
        "github.com/json-iterator/go": {
            "name": "com_github_json_iterator_go",
            "commit": "f2b4162afba35581b6d4a50d3b8f34e33c144682",
            "importpath": "github.com/json-iterator/go",
        },
        "github.com/mailru/easyjson": {
            "name": "com_github_mailru_easyjson",
            "commit": "2f5df55504ebc322e4d52d34df6a1f5b503bf26d",
            "importpath": "github.com/mailru/easyjson",
        },
        "github.com/pborman/uuid": {
            "name": "com_github_pborman_uuid",
            "commit": "ca53cad383cad2479bbba7f7a1a05797ec1386e4",
            "importpath": "github.com/pborman/uuid",
        },
        "github.com/peterbourgon/diskv": {
            "name": "com_github_peterbourgon_diskv",
            "commit": "5f041e8faa004a95c88a202771f4cc3e991971e6",
            "importpath": "github.com/peterbourgon/diskv",
        },
        "github.com/PuerkitoBio/purell": {
            "name": "com_github_puerkitobio_purell",
            "commit": "8a290539e2e8629dbc4e6bad948158f790ec31f4",
            "importpath": "github.com/PuerkitoBio/purell",
        },
        "github.com/PuerkitoBio/urlesc": {
            "name": "com_github_puerkitobio_urlesc",
            "commit": "5bd2802263f21d8788851d5305584c82a5c75d7e",
            "importpath": "github.com/PuerkitoBio/urlesc",
        },
        # also depended upon by github.com/gin-gonic/gin
        "github.com/ugorji/go": {
            "name": "com_github_ugorji_go",
            "commit": "ded73eae5db7e7a0ef6f55aace87a2873c5d2b74",
            "importpath": "github.com/ugorji/go",
        },
        "golang.org/x/time": {
            "name": "org_golang_x_time",
            "commit": "f51c12702a4d776e4c1fa9b0fabab841babae631",
            "importpath": "golang.org/x/time",
        },
#        "golang.org/x/tools": {
#            "name": "org_golang_x_tools",
#            "commit": "2382e3994d48b1d22acc2c86bcad0a2aff028e32",
#            "importpath": "golang.org/x/tools",
#        },
        "gopkg.in/inf.v0": {
            "name": "in_gopkg_inf_v0",
            "commit": "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
            "importpath": "gopkg.in/inf.v0",
        },
        # also depended upon by github.com/gin-gonic/gin
        "gopkg.in/yaml.v2": {
            "name": "in_gopkg_yaml_v2",
            "commit": "670d4cfef0544295bc27a114dbac37980d83185a",
            "importpath": "gopkg.in/yaml.v2",
        },
        "k8s.io/apiserver": {
            "name": "io_k8s_apiserver",
            "tag": "kubernetes-1.11.0",
            "importpath": "k8s.io/apiserver",
        },
        "k8s.io/kube-openapi": {
            "name": "io_k8s_kube_openapi",
            "commit": "91cfa479c814065e420cee7ed227db0f63a5854e",
            "importpath": "k8s.io/kube-openapi",
        },
        "github.com/modern-go/reflect2": {
            "name": "com_github_modern_go_reflect2",
            "commit": "05fbef0ca5da472bbf96c9322b84a53edc03c9fd",
            "importpath": "github.com/modern-go/reflect2",
        },
        "github.com/modern-go/concurrent": {
            "name": "com_github_modern_go_concurrent",
            "commit": "bacd9c7ef1dd9b15be4a9909b8ac7a4e313eec94",
            "importpath": "github.com/modern-go/concurrent",
        }
    },
}

def go_dependencies():
  for go_dependencies in GO_DEPENDENCIES.values():
    for dependency in go_dependencies.values():
      go_repository(**dependency)
