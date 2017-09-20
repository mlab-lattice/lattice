package main

import (
	"flag"
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crdclient "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	corev1 "k8s.io/api/core/v1"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

var (
	kubeconfigPath string
	providerName   string
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&providerName, "provider", "", "path to kubeconfig file")
	flag.Parse()
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

	kubeClientset := clientset.NewForConfigOrDie(kubeconfig)
	latticeInternalNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.InternalNamespace,
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.CoreV1().Namespaces().Create(latticeInternalNamespace)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})

	latticeUserNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(coreconstants.UserSystemNamespace),
		},
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.CoreV1().Namespaces().Create(latticeUserNamespace)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})

	_, err = crdclient.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		panic(err)
	}

	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	var buildConfig crv1.ComponentBuildConfig
	switch coretypes.Provider(providerName) {
	case coreconstants.ProviderLocal:
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
		panic("unsupported providerName")
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
