package main

import (
	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func seedConfig(kubeconfig *rest.Config) {
	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	var dockerRegistry string
	if dev {
		dockerRegistry = localDevDockerRegistry
	} else {
		dockerRegistry = devDockerRegistry
	}

	// Create config
	var buildConfig crv1.ComponentBuildConfig
	var envoyConfig crv1.EnvoyConfig
	switch coretypes.Provider(providerName) {
	case coreconstants.ProviderLocal:
		buildConfig = crv1.ComponentBuildConfig{
			DockerConfig: crv1.BuildDockerConfig{
				Registry:           "lattice-local",
				RepositoryPerImage: true,
				Push:               false,
			},
			PullGitRepoImage: dockerRegistry + "/pull-git-repo",
			BuildDockerImage: dockerRegistry + "/build-docker-image",
		}

		envoyConfig = crv1.EnvoyConfig{
			PrepareImage:      dockerRegistry + "prepare-envoy",
			Image:             "lyft/envoy",
			EgressPort:        9001,
			RedirectCidrBlock: "172.16.29.0/16",
			XdsApiPort:        8080,
		}
	default:
		panic("unsupported providerName")
	}

	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: "global",
		},
		Spec: crv1.ConfigSpec{
			ComponentBuild: buildConfig,
			Envoy:          envoyConfig,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return nil, crClient.Post().
			Namespace("default").
			Resource(crv1.ConfigResourcePlural).
			Body(config).
			Do().Into(nil)
	})
}
