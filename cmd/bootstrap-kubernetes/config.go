package main

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func seedConfig(kubeconfig *rest.Config, userSystemUrl, systemIP string) {
	fmt.Println("Seeding lattice config...")
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
				PrepareImage:      dockerRegistry + "/prepare-envoy",
				Image:             "lyft/envoy",
				RedirectCidrBlock: "172.16.29.0/16",
				XdsApiPort:        8080,
			},
			Provider: provider,
		},
	}

	buildConfig := crv1.ComponentBuildConfig{
		DockerConfig: crv1.BuildDockerConfig{
			RepositoryPerImage: true,
			Push:               true,
		},
		PullGitRepoImage: dockerRegistry + "/pull-git-repo",
		BuildDockerImage: dockerRegistry + "/build-docker-image",
	}

	switch provider {
	case coreconstants.ProviderLocal:
		buildConfig.DockerConfig.Push = false
		// TODO: is there a better value for this? are there any security
		// 		 implications here? what if we own the lattice-local docker hub repo?
		buildConfig.DockerConfig.Registry = "lattice-local"
		config.Spec.SystemIP = &systemIP
	default:
		panic("unsupported provider")
	}

	config.Spec.ComponentBuild = buildConfig

	pollKubeResourceCreation(func() (interface{}, error) {
		return nil, crClient.Post().
			Namespace(constants.NamespaceLatticeInternal).
			Resource(crv1.ConfigResourcePlural).
			Body(config).
			Do().Into(nil)
	})
}
