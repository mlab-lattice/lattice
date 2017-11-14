package main

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crdclient "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func seedConfig(kubeconfig *rest.Config, userSystemUrl, systemIP, region string) {
	fmt.Println("Seeding lattice config...")
	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create config
	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ConfigGlobal,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: crv1.ConfigSpec{
			SystemConfigs: map[coretypes.LatticeNamespace]crv1.SystemConfig{
				coreconstants.UserSystemNamespace: {
					Url: userSystemUrl,
				},
			},
			Envoy: crv1.EnvoyConfig{
				PrepareImage:      latticeContainerRegistry + "/envoy-prepare-envoy",
				Image:             "envoyproxy/envoy",
				RedirectCidrBlock: "172.16.29.0/16",
				XdsApiPort:        8080,
			},
			ComponentBuild: crv1.ComponentBuildConfig{
				DockerConfig: crv1.BuildDockerConfig{
					RepositoryPerImage: false,
					Repository:         constants.DockerRegistryComponentBuildsDefault,
					Push:               true,
					Registry:           componentBuildRegistry,
				},
				BuildDockerImage: latticeContainerRegistry + "/component-build-build-docker-image",
				GetEcrCredsImage: latticeContainerRegistry + "/component-build-get-ecr-creds",
				PullGitRepoImage: latticeContainerRegistry + "/component-build-pull-git-repo",
			},
		},
	}

	switch provider {
	case coreconstants.ProviderLocal:
		config.Spec.ComponentBuild.DockerConfig.Push = false
		config.Spec.ProviderConfig.Local = &crv1.ProviderConfigLocal{
			IP: systemIP,
		}
	case coreconstants.ProviderAWS:
		config.Spec.ProviderConfig.AWS = &crv1.ProviderConfigAWS{
			Region: region,
		}
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return nil, crClient.Post().
			Namespace(constants.NamespaceLatticeInternal).
			Resource(crv1.ConfigResourcePlural).
			Body(config).
			Do().Into(nil)
	})
}
