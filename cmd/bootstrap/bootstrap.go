package main

import (
	"flag"
	"fmt"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	providerutil "github.com/mlab-lattice/core/pkg/provider"

	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"github.com/golang/glog"
)

var (
	kubeconfigPath string
	provider       string
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&provider, "provider", "", "path to kubeconfig file")
	flag.Parse()

	if !providerutil.ValidateProvider(provider) {
		panic(fmt.Sprintf("Invalid provider %v", provider))
	}
}

func main() {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(kubeconfig)
	if err != nil {
		panic(err)
	}

	_, err = crdclient.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		panic(err)
	}

	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	var buildConfig crv1.ComponentBuildConfig
	switch provider {
	case providerutil.Local:
		buildConfig = crv1.ComponentBuildConfig{
			DockerConfig: crv1.BuildDockerConfig{
				Registry:           "lattice-local",
				RepositoryPerImage: true,
				Push:               false,
			},
			PullGitRepoImage: "bazel/docker:pull-git-repo",
			BuildDockerImage: "bazel/docker:build-docker-image",
		}
	default:
		panic("unsupported provider")
	}

	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: "global",
		},
		Spec: crv1.ConfigSpec{
			ComponentBuild: buildConfig,
		},
	}

	err = crClient.Post().
		Namespace("default").
		Resource(crv1.ConfigResourcePlural).
		Body(config).
		Do().Into(nil)

	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			panic(err)
		}

		glog.Warning("Config already exists")
	}
}
